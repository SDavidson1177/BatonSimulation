# Baton Simulator

This program simulates messages being sent using the Baton interoperability protocol. It also simulates sending IBC packets between chain pairs using multi-hop IBC channels, only single-hop IBC channels and hub blockchains.

## Usage

```BASH
main.go [edges csv file] [channel_type] [send interval] [jitter] [number of sends] [direct] [hubs...]
```

## Instructions

### CSV File

The csv file should be a list of blockchain pairs (integer IDs) where each pair represents an IBC connection.

**example**

```CSV
1,2
2,3
3,1
```

### Channel Type

When set to 'multi', the simulator will allow indirectly connected blockchains to communicate. Only light client updates will be submitted to intermediate blockchains along a route. 

When set to 'single', only single-hop channels can be used. If a route consists of multiple hops, a packet will be delievered at each hop. Furthremore, the simulator will wait before sending on the next hop, since one block height must pass before the packet can be transmitted again.

### Send Interval

The minimum amount of milliseconds between subsequent sends for any given blockchain pair.

### Jitter

The send time between packets is equal to...

```
send_interval + random_in_range(0, jitter)
```

### Number of Sends

The total number of packets to simulate

### Direct

Can be true or false. If true, only allow blockchain pairs to communicate if they are directly connected, or connected via a sequence of hub blockchains. If false, allow all indirectly connected blockchains to communicate.

### Hubs

List in the following format

```
baton-1 baton-2 baton-3...
```

The example gives blockchains with IDs 1, 2 and 3 from csv file. Listed chains are treated as hub blockchains.

## Output

A log of send, deliver and client updates events are given as output. The maximum number of transactions in any given block and the total number of transactions is given for each blockchain.

