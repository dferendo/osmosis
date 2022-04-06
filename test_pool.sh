NODE="tcp://localhost:26657"
CHAIN_ID="osmo-1"
NODE_DATA="./data/$CHAIN_ID/"

# Run the commands one-by-one in a bash
starport chain serve --config=config.yml --reset-once

osmosisd query bank balances osmo1rgy3fw5z64mn8k45xpg9vxk9d90e9lj68lmq2u --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"

# Create pools
osmosisd tx gamm create-pool --pool-file="./pools/pool_uband.json" --fees="350stake" --from=alice --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"
osmosisd tx gamm create-pool --pool-file="./pools/pool_ukava.json" --fees="350stake" --from=alice --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"
osmosisd tx gamm create-pool --pool-file="./pools/pool_uluna.json" --fees="350stake" --from=alice --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"

# Testing, no need to execute
osmosisd tx gamm swap-exact-amount-in 10000stake 10 --fees=1750stake --from=alice --swap-route-pool-ids=1 --swap-route-denoms=stake --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"
osmosisd query bank balances osmo16e20tp8zw2zu5z57nnzzhl65m7499uaqq933yr --node="$NODE" --chain-id="$CHAIN_ID" --home="$NODE_DATA"

