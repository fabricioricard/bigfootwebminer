package banmgr

import (
	"net"
	"sync"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/pktlog/log"
)

type Config struct {
	DisableBanning bool
	IpWhiteList    []string
	BanThreashold  uint32
}

type BanInfo struct {
	addr           string
	reason         string
	banScore       uint32
	banExpiresTime time.Time
}

type BannedPeers struct {
	time   time.Time
	reason string
}

type SuspiciousPeers struct {
	banReason       *string
	dynamicBanScore dbs
}

type dbs struct {
	bs          *DynamicBanScore
	lastUsedSec int64
}

type BanMgr struct {
	config     *Config
	banned     map[string]BannedPeers
	m          sync.Mutex
	suspicious map[string]SuspiciousPeers
}

func now() int64 {
	return time.Now().Unix()
}

func TrimAddress(host string) string {
	address, _, err := net.SplitHostPort(host)
	if err != nil {
		log.Debugf("can't split hostport %v", err)
		return address
	}
	return address
}

func New(config *Config) *BanMgr {
	c := &Config{
		DisableBanning: config.DisableBanning,
		IpWhiteList:    config.IpWhiteList,
		BanThreashold:  config.BanThreashold,
	}

	return &BanMgr{
		config:     c,
		suspicious: make(map[string]SuspiciousPeers),
		banned:     make(map[string]BannedPeers),
	}
}

func (b *BanMgr) IsBanned(ip string) bool {
	addr := TrimAddress(ip)

	if banned, ok := b.banned[addr]; ok {
		if time.Now().Before(banned.time) {
			log.Debugf("Peer %s is banned for another %v - disconnecting", addr, time.Until(banned.time))
			return true
		}
		log.Infof("Peer %s is no longer banned", addr)
		delete(b.banned, addr)
	}
	return false
}

func (b *BanMgr) ForEachIp(f func(bi BanInfo)) er.R {
	//TODO: lock copy and unlock
	b.m.Lock()
	var toRemove []string
	for ip, peer := range b.banned {
		if !time.Now().Before(peer.time) {
			toRemove = append(toRemove, ip)
		}
	}
	//TODO: what about suspicious peers?
	b.m.Unlock()

	for _, addr := range toRemove {
		delete(b.banned, addr)
	}
	return nil
}

func (b *BanMgr) AddBanScore(ip string, persistent, transient uint32, reason string) bool {
	// No warning is logged and no score is calculated if banning is disabled.
	if b.config.DisableBanning {
		return false
	}

	for _, item := range b.config.IpWhiteList {
		if item == ip {
			log.Debugf("Misbehaving whitelisted peer %s: %s", ip, reason)
			return false
		}
	}

	if b.suspicious == nil {
		log.Debugf("Misbehaving peer %s: %s and no ban manager yet")
		return false
	}

	warnThreshold := b.config.BanThreashold >> 1
	if transient == 0 && persistent == 0 {
		// The score is not being increased, but a warning message is still
		// logged if the score is above the warn threshold.
		score := b.suspicious[ip].dynamicBanScore.bs.Int()
		if score > warnThreshold {
			log.Warnf("Misbehaving peer %s: %s -- ban score is %d, it was not increased this time", ip, reason, score)
		}
		return false
	}
	b.m.Lock()
	score := b.suspicious[ip].dynamicBanScore.bs.Increase(persistent, transient)
	b.m.Unlock()
	if score > warnThreshold {
		log.Warnf("Misbehaving peer %s: %s -- ban score increased to %d", ip, reason, score)
		if score > b.config.BanThreashold {
			log.Warnf("Misbehaving peer %s -- banning and disconnecting", ip)
			//add to banned
			b.banned[ip] = BannedPeers{time.Now(), reason}
			return true
			//Will be done by the server
			//sp.server.BanPeer(ip)
			//sp.Disconnect()
		}
	}
	return false
}

func (b *BanMgr) GetScore(host string) *DynamicBanScore {
	b.m.Lock()
	host = TrimAddress(host)
	if _, ok := b.suspicious[host]; !ok {
		log.Debugf("Create new banScore for [%s]", host)
		if b.suspicious == nil {
			b.suspicious = make(map[string]SuspiciousPeers)
		}
		b.suspicious[host] = SuspiciousPeers{
			banReason: nil,
			dynamicBanScore: dbs{
				bs:          &DynamicBanScore{},
				lastUsedSec: now(),
			},
		}
	}
	bs := b.suspicious[host].dynamicBanScore.bs
	b.m.Unlock()
	return bs
}
