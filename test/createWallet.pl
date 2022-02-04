#!  /usr/bin/perl -w

################################################################################
#   requires the installation of cpan, and after needs to run the following:
#   $ echo "install Expect" | cpan
################################################################################

use strict;
use Expect;

my  $PLDCTL='./bin/pldctl';
my  $PLDCTL_ARGS='create';
my  $WALLET_PASSWORD='w4ll3tP@sswd';
my  $SEED_PASSPHRASE='s3edP@ssphr4z';

#   check cli arguments
my  $VERBOSE = '';

if( $#ARGV + 1 == 1 ) {
    $VERBOSE = $ARGV[ 0 ];
}

#   open a pipe to interact with pldctl
my  $pldctl = Expect->new;
my  $inputLine;

$pldctl->raw_pty( 1 );
$pldctl->spawn( qq/$PLDCTL $PLDCTL_ARGS/ )
    or die qq/Failed to open pipe to $PLDCTL: $!/;

if( $VERBOSE eq '--verbose' ) {
    $pldctl->exp_internal( 1 );
}

#   wait pldctl ask for wallet password, and the send it
$pldctl->expect( 10, 'Input wallet password: ' );
$pldctl->send( qq/$WALLET_PASSWORD\n/ );

#   wait pldctl ask for wallet password confirmation, and the send it
$pldctl->expect( 10, 'Confirm password: ' );
$pldctl->send( qq/$WALLET_PASSWORD\n/ );

#   wait pldctl ask for seed existence, and the send it
$pldctl->expect( 10, 'Do you have an existing Pktwallet seed you want to use? (Enter y\/n): ' );
$pldctl->send( qq/n\n/ );

#   wait pldctl ask for cipher seed passphrase, and the send it
$pldctl->expect( 10, 'Input your passphrase if you wish to encrypt it (or press enter to proceed without a cipher seed passphrase): ' );
$pldctl->send( qq/$SEED_PASSPHRASE\n/ );

#   wait pldctl ask for cipher seed passphrase confirmation, and the send it
$pldctl->expect( 10, 'Confirm password: ' );
$pldctl->send( qq/$SEED_PASSPHRASE\n/ );

#   wait pldctl shows a successful message
$pldctl->expect( 10, 'pld successfully initialized!' );
