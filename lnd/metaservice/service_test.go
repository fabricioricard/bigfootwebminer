////////////////////////////////////////////////////////////////////////////////
//	lnd/metaservice/service_test.go  -  Jan-26-2022  -  aldebap
//
//	unit tests for lnd/metaservice/service.go
////////////////////////////////////////////////////////////////////////////////

package metaservice

import (
	"context"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/btcutil/util"
	"github.com/pkt-cash/pktd/chaincfg"
	"github.com/pkt-cash/pktd/lnd/channeldb/kvdb"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnwallet/btcwallet"
	"github.com/pkt-cash/pktd/lnd/macaroons"
	"github.com/pkt-cash/pktd/neutrino"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/snacl"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet"
	"github.com/pkt-cash/pktd/pktwallet/walletdb"
	"github.com/stretchr/testify/require"
)

const (
	testWalletFilename string = "wallet.db"
)

var (
	testPassword = []byte("test-password")
	testSeed     = []byte("test-seed-123456789")
	testMac      = []byte("fakemacaroon")

	testNetParams = &chaincfg.MainNetParams

	defaultRootKeyIDContext = macaroons.ContextWithRootKeyID(
		context.Background(), macaroons.DefaultRootKeyID,
	)
)

//	Test that as error occurs on an attempt to change the password for a non-existing  wallet
func TestChangePasswordForNonExistingWallet(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestChangePasswordForNonExistingWallet()")

	//	create a temporary directory, initialize an empty walletdb with an SPV chain
	//	namespace, and create a configuration for the ChainService
	testDir, err := ioutil.TempDir("", "neutrino")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	db, errr := walletdb.Create("bdb", testDir+"/testNeutrino.db", true)
	util.RequireNoErr(t, errr)
	defer db.Close()

	config := neutrino.Config{
		DataDir:     testDir,
		Database:    db,
		ChainParams: *testNetParams,
	}

	testChainService, errr := neutrino.NewChainService(config)
	util.RequireNoErr(t, errr)

	//	create a new MetaService with our test file
	metaService := NewMetaService(testChainService)
	metaService.walletPath = btcwallet.NetworkDir(testDir, testNetParams)
	metaService.walletFile = testWalletFilename

	ctx := context.Background()

	//	changing the password to a non-existing wallet should fail
	log.Debugf(">>>>> [1] attempt to change password for a non-existing wallet")

	newPassword := []byte("hunter2???")
	req := &lnrpc.ChangePasswordRequest{
		CurrentPassword:    testPassword,
		CurrentPubPassword: testPassword,
		NewPassword:        newPassword,
		NewMacaroonRootKey: true,
	}

	_, err = metaService.ChangePassword(ctx, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wallet not found")
}

//	Test that we can successfully change the wallet's password needed to unlock
//	it and rotate the root key for the macaroons in the same process.
func TestChangeWalletPasswordNewRootkey(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestChangeWalletPasswordNewRootkey()")

	//	create a temporary directory for the macaroon files
	//	TODO: metaservice.changePassword() is not async but, it's waiting on a channel for
	//		the macaroon response so, it's not possible to test macaroon stuff anymore
	/*
		macTestDir, err := ioutil.TempDir("", "macaroon")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(macTestDir)
		}()

		// Changing the password of the wallet will also try to change the
		// password of the macaroon DB. We create a default DB here but close it
		// immediately so the service does not fail when trying to open it.
		store, errr := openOrCreateTestMacStore(macTestDir, &testPassword, testNetParams)
		util.RequireNoErr(t, errr)

		errr = store.Close()
		util.RequireNoErr(t, errr)

		// Create some files that will act as macaroon files that should be
		// deleted after a password change is successful with a new root key
		// requested.
		var macTempFiles []string

		for i := 0; i < 3; i++ {
			file, err := ioutil.TempFile(macTestDir, "")
			require.NoError(t, err)

			macTempFiles = append(macTempFiles, file.Name())
			err = file.Close()
			require.NoError(t, err)

			log.Debugf(">>>>> [1] macaroon file created: %s", file.Name())
		}
	*/

	//	create a temporary directory, initialize an empty walletdb with an SPV chain
	//	namespace, and create a configuration for the ChainService
	testDir, err := ioutil.TempDir("", "neutrino")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(testDir)
	}()

	walletDb, errr := walletdb.Create("bdb", testDir+"/testNeutrino.db", true)
	util.RequireNoErr(t, errr)
	defer walletDb.Close()

	/*
		neutrinoDb, errr := headerfs.NewNeutrinoDBStore(walletDb, &chaincfg.MainNetParams, true)
		util.RequireNoErr(t, errr)
		_ = neutrinoDb
	*/

	config := neutrino.Config{
		DataDir:     testDir,
		Database:    walletDb,
		ChainParams: *testNetParams,
	}
	testChainService, errr := neutrino.NewChainService(config)
	util.RequireNoErr(t, errr)
	//	testChainService.Start()
	//	defer testChainService.Stop()

	//	create a new MetaService with our test file
	metaService := NewMetaService(testChainService)
	metaService.walletPath = btcwallet.NetworkDir(testDir, testNetParams)
	metaService.walletFile = testWalletFilename

	/*
		metaService.netParams = testNetParams
		metaService.chainDir = macTestDir
		metaService.macaroonFiles = macTempFiles
	*/

	ctx := context.Background()

	// Create a wallet to test changing the password.
	createTestWallet(t, testDir, testNetParams)

	//	changing the wallet's password using an incorrect current password should fail
	newPassword := []byte("hunter2???")

	log.Debugf(">>>>> [3] attempt to change passwaord using an incorrect current password")
	wrongReq := &lnrpc.ChangePasswordRequest{
		CurrentPassword: []byte("wrong-ofc"),
		NewPassword:     newPassword,
	}
	_, err = metaService.ChangePassword(ctx, wrongReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid passphrase for master public key")

	//	macaroon files must still exist after an unsuccessful attempt to change password
	/*
		for _, tempFile := range macTempFiles {
			_, err := os.Stat(tempFile)
			require.False(t, os.IsNotExist(err), "missing macaroon file: %s", tempFile)
			log.Debugf(">>>>> [4] macaroon file still exists: %s", tempFile)
		}
	*/

	//	changing the wallet's password using an invalid new password should fail
	log.Debugf(">>>>> [5] attempt to change passwaord using an invalid new password")

	wrongReq.NewPassword = []byte("8")
	_, err = metaService.ChangePassword(ctx, wrongReq)
	require.Error(t, err)
	require.Contains(t, err.Error(), "custom password must have at least 8 characters")

	//	when providing the correct wallet's current password and a valid new password,
	//	the password change should succeed
	log.Debugf(">>>>> [6] finally change passwaord")

	req := &lnrpc.ChangePasswordRequest{
		CurrentPassword:    testPassword,
		CurrentPubPassword: testPassword,
		NewPassword:        newPassword,
		NewMacaroonRootKey: true,
	}

	_, errr = changePassword(metaService, testDir, req)
	util.RequireNoErr(t, errr)

	//	macaroon files must must no exist
	/*
		for _, tempFile := range macTempFiles {
			_, err := os.Stat(tempFile)
			require.True(t, os.IsNotExist(err), "macaroon file should not exist: %s", tempFile)
		}
	*/
}

// TestChangeWalletPasswordStateless checks that trying to change the password
// of an existing wallet that was initialized stateless works when when the
// --stateless_init flat is set. Also checks that if no password is given,
// the default password is used.
func TestChangeWalletPasswordStateless(t *testing.T) {
	t.Parallel()

	log.Debugf(">>>>> running TestChangeWalletPasswordStateless()")

	//	TODO: metaservice.changePassword() is not async but, it's waiting on a channel for
	//		the macaroon response so, it's not possible to test macaroon stuff anymore
	/*
		//	create a temporary directory, initialize an empty walletdb with an SPV chain
		//	namespace, and create a configuration for the ChainService
		testDir, err := ioutil.TempDir("", "stateless-neutrino")
		require.NoError(t, err)
		defer func() {
			_ = os.RemoveAll(testDir)
		}()

		// Changing the password of the wallet will also try to change the
		// password of the macaroon DB. We create a default DB here but close it
		// immediately so the service does not fail when trying to open it.
		store, errr := openOrCreateTestMacStore(
			testDir, &lnwallet.DefaultPrivatePassphrase, testNetParams,
		)
		util.RequireNoErr(t, errr)
		util.RequireNoErr(t, store.Close())

		// Create a temp file that will act as the macaroon DB file that will
		// be deleted by changing the password.
		tmpFile, err := ioutil.TempFile(testDir, "")
		require.NoError(t, err)
		tempMacFile := tmpFile.Name()
		err = tmpFile.Close()
		require.NoError(t, err)

		// Create a file name that does not exist that will be used as a
		// macaroon file reference. The fact that the file does not exist should
		// not throw an error when --stateless_init is used.
		nonExistingFile := path.Join(testDir, string(testMac))

		walletDb, errr := walletdb.Create("bdb", testDir+"/testNeutrino.db", true)
		util.RequireNoErr(t, errr)
		defer walletDb.Close()

		config := neutrino.Config{
			DataDir:     testDir,
			Database:    walletDb,
			ChainParams: *testNetParams,
		}
		testChainService, errr := neutrino.NewChainService(config)
		util.RequireNoErr(t, errr)

		//	create a new MetaService with our test file
		metaService := NewMetaService(testChainService)
		metaService.walletPath = btcwallet.NetworkDir(testDir, testNetParams)
		metaService.walletFile = testWalletFilename

		metaService.netParams = testNetParams
		metaService.chainDir = testDir
		metaService.macaroonFiles = []string{tempMacFile, nonExistingFile}

		ctx := context.Background()

		// Create a wallet we can try to unlock. We use the default password
		// so we can check that the unlocker service defaults to this when
		// we give it an empty CurrentPassword to indicate we come from a
		// --noencryptwallet state.
		createTestWalletWithPw(
			t, lnwallet.DefaultPublicPassphrase,
			lnwallet.DefaultPrivatePassphrase, testDir, testNetParams,
		)

		// We make sure that we get a proper error message if we forget to
		// add the --stateless_init flag but the macaroon files don't exist.
		log.Debugf(">>>>> [3] attempt to change passwaord without --stateless_init flag")

		badReq := &lnrpc.ChangePasswordRequest{
			NewPassword:        testPassword,
			NewMacaroonRootKey: true,
		}
		_, err = metaService.ChangePassword(ctx, badReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "if the wallet was initialized stateless")

		// Prepare the correct request we are going to send to the unlocker
		// service. We don't provide a current password to indicate there
		// was none set before.
		log.Debugf(">>>>> [4] finally change passwaord")

		req := &lnrpc.ChangePasswordRequest{
			NewPassword:        testPassword,
			StatelessInit:      true,
			NewMacaroonRootKey: true,
		}

		_, errr = changePassword(metaService, testDir, req)
		util.RequireNoErr(t, errr)
	*/
}

//	execute a password change
func changePassword(metaService *MetaService, macTestDir string, req *lnrpc.ChangePasswordRequest) (*lnrpc.ChangePasswordResponse, er.R) {

	//	when providing the correct wallet's current password and a valid new password,
	//	the password change should succeed
	ctx := context.Background()
	response, err := metaService.ChangePassword(ctx, req)
	if err != nil {
		return nil, er.Errorf("could not change password: %v", err)
	}

	/*
		if !bytes.Equal(response.AdminMacaroon, testMac) {
			return nil, er.Errorf("mismatched macaroon: expected %x, got %x", testMac, response.AdminMacaroon)
		}
	*/
	//	close the macaroon DB and try to open it and read the root key with the
	//	new password
	store, errr := openOrCreateTestMacStore(macTestDir, &testPassword, testNetParams)
	if errr != nil {
		return nil, er.Errorf("could not create test store: %v", err)
	}
	_, _, err = store.RootKey(defaultRootKeyIDContext)
	if err != nil {
		return nil, er.Errorf("could not get root key: %v", errr)
	}
	errr = store.Close()
	if errr != nil {
		return nil, er.Errorf("could not close store: %v", err)
	}

	// Do cleanup now. Since we are in a go func, the defer at the
	// top of the outer would not work, because it would delete
	// the directory before we could check the content in here.
	err = os.RemoveAll(macTestDir)
	if errr != nil {
		return nil, er.E(err)
	}

	return response, nil
}

// openOrCreateTestMacStore opens or creates a bbolt DB and then initializes a
// root key storage for that DB and then unlocks it, creating a root key in the
// process.
func openOrCreateTestMacStore(tempDir string, pw *[]byte,
	netParams *chaincfg.Params) (*macaroons.RootKeyStorage, er.R) {

	netDir := btcwallet.NetworkDir(tempDir, netParams)
	errr := os.MkdirAll(netDir, 0700)
	if errr != nil {
		return nil, er.E(errr)
	}
	db, err := kvdb.Create(
		kvdb.BoltBackendName, path.Join(netDir, macaroons.DBFilename),
		true,
	)
	if err != nil {
		return nil, err
	}

	store, err := macaroons.NewRootKeyStorage(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	err = store.CreateUnlock(pw)
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	_, _, errr = store.RootKey(defaultRootKeyIDContext)
	if errr != nil {
		_ = store.Close()
		return nil, er.E(errr)
	}

	return store, nil
}

func createTestWallet(t *testing.T, dir string, netParams *chaincfg.Params) {
	createTestWalletWithPw(t, testPassword, testPassword, dir, netParams)
}

func createTestWalletWithPw(t *testing.T, pubPw, privPw []byte, dir string, netParams *chaincfg.Params) {

	// Instruct waddrmgr to use the cranked down scrypt parameters when
	// creating new wallet encryption keys.
	fastScrypt := waddrmgr.FastScryptOptions
	keyGen := func(passphrase *[]byte, config *waddrmgr.ScryptOptions) (
		*snacl.SecretKey, er.R) {

		return snacl.NewSecretKey(
			passphrase, fastScrypt.N, fastScrypt.R, fastScrypt.P,
		)
	}
	waddrmgr.SetSecretKeyGen(keyGen)

	// Create a new test wallet that uses fast scrypt as KDF.
	netDir := btcwallet.NetworkDir(dir, netParams)
	loader := wallet.NewLoader(netParams, netDir, testWalletFilename, true, 0)
	_, err := loader.CreateNewWallet(
		pubPw, privPw, []byte(hex.EncodeToString(testSeed)), time.Time{}, nil,
	)
	util.RequireNoErr(t, err)

	realWalletPathname := wallet.WalletDbPath(netDir, testWalletFilename)
	log.Debugf(">>>>> [1] wallet path: %s", realWalletPathname)
	walletFileExists := true
	_, errr := os.Stat(realWalletPathname)
	if err != nil {
		if os.IsNotExist(errr) {
			walletFileExists = false
		} else {
			require.NoError(t, errr)
		}
	}
	log.Debugf(">>>>> [2] after loader.CreateNewWallet() the wallet file exists: %t", walletFileExists)

	err = loader.UnloadWallet()
	util.RequireNoErr(t, err)
}
