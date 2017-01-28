# siapool


[![Build Status](https://travis-ci.org/siapool/p2pool.svg?branch=master)](https://travis-ci.org/siapool/p2pool)

## Status

Early development phase, completely useless at the moment.

The intention is to make a p2pool for SIA. In a first phase the pool interface and blockgeneration will be created. This will result in a fully functional but centralized pplns pool. The sharechain is currently just a list of accepted shares and the p2peer protocol will be added in phase 2.

## Connect your miner

Only stratum as defined on https://siamining.com/stratum is supported, no `getHeader` implementations.
Direct your miner to the pool using the following host: `<poolhost>:<poolport>`

Example using gominer:
```
gominer -url tcp+stratum://siapool.tech:3333 -user 1e80b18e7cdd92c3a03f307c5f453bb5a26784dfce054063b4976c8784b3a98f55ecf5f59627
```

The benefit of using stratum is that the server does not need to store all generated headers since the clients generate the randomness. This makes the server much cleaner, more lightweight and enables it to support a lot more miners. The major drawback is that official Sia gpu miner is not compatible.

## Share difficulty

The pool has a starting difficulty for a 1Gh/s miner to find two shares/day on average. Target pool wide sharetime is 30 seconds and the length of the sharechain is 2 * 1440 * 4 (= 4 days). The difficulty of the pool is adjusted every 100 shares and calculated over the entire sharechain.

## Payout logic

Each share contains a generation transaction that pays to the previous n shares, where n is the length of the sharechain.

The block reward and the transaction fees are combined and apportioned according to these rules:

A subsidy of 0.5% is sent to the miner that solved the block in order to discourage not sharing solutions that qualify as a block. (A miner with the aim to harm others could withhold the block, thereby preventing anybody from getting paid. He can NOT redirect the payout to himself.) The remaining 99.5% is distributed evenly to miners based on work done recently.

A node can choose to keep a fee for operating the node.

In the event that a share qualifies as a block, this generation transaction is exposed to the Sia network and takes effect, transferring each miner its payout.


## Architectural concept

Siapool needs a lot of information from the sia network to be able to construct the blocks for which it hands out headers to miners and needs to feed complete blocks to the sia network. Siad does not expose this information through it's api and siapool needs to react fast on new blocks. It's a lot more comfortable if siapool accesses the internal datastructures of siad directly to be able to serve it's miners up to date jobs and to submit custom made blocks to the sia network.

This left the option of implementing siapool as a siad module or vice versa, namely importing siad and launching the modules we require ourselves. The second option has been chosen to limit the impact on the sia project itself and to leave the pool landscape for sia mining open.

An additional benefit of embedding the necessary siad functionality is that there is only a single binary, there is no need to run a separate siad and to configure the pool and siad to work together.

## How to

* **How to check the state of the embedded siad (synchronization, peers, ...)?**

  The embedded siad's api is exposed on `localhost:9980`, the same as a normal siad. This means you can use the normal siac commandline utility to talk to it. Only the consensus, transactionpool and gateway modules are loaded so wallet operations will not work.
