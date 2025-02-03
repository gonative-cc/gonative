#!/usr/bin/env fish

if test (count $argv) -ne 1
    echo "Usage: " $(status -f | xargs basename) "<path to gonaitve>"
    exit 1
end

set APP $argv[1]
if test -z (command -s $APP)
    echo "binary " $APP " not found"
    exit 2
end

echo ">>> app binary:" $APP
# killall "$APP" &>/dev/null || true


# chain id flag:
set cid native-t1

set DENOM untiv
set SCALE_FACTOR 000000
set NATIVEBASE {$SCALE_FACTOR}untiv

set devnetcoins "1000000$NATIVEBASE"
set valcoins "10$NATIVEBASE"
set valdelegation "1$NATIVEBASE"

set val native1z6v80kr6quj7gwndy5h2c4la9ufc0lu90svy3g
set val_name val-d1
set valop

set devnet native1n7awwqdvwcuey8rl84kuypmg9hhp690y246k5e

# $APP keys add $val_name --recover
#    change ski lab lab wise stomach allow poem witness stairs chief result dirt sign junk water interest hover budget hurdle before same rocket retreat
#  devnet:
#    evoke body truly crisp member adapt section outer coyote hunt resist bomb divorce daring pill crush foster act walk melt smoke venue nothing just
# yes "$VAL1_MNEMONIC\n" | $APP keys add $VAL0_KEY --recover

rm -rf ~/.gonative
$APP init bee-t1 --chain-id $cid
sleep 0.2
node ./testnet.js
sleep 0.2

$APP genesis add-genesis-account $val $valcoins
$APP genesis add-genesis-account $devnet $devnetcoins

$APP genesis gentx $val_name $valdelegation --chain-id $cid
$APP genesis collect-gentxs
$APP genesis validate

echo ">>> init chain:" $APP init bee-t1 --chain-id $cid
$APP start
