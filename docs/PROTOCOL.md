# Protocol overview

> [!NOTE]  
> Current version is a proof of concept. It has some obvious limitations and drawbacks.

Version 1 was implemented in such way for the simplicity sake of the first version.
The goal was to make a simple solution first and then address the issues in Version 2.

## Encryption

All messages in the room are encrypted with a single symmetric key. There's no other encryption or signing.

This prevents other users in the network to read your room's messages.
Yet this also means that any message can be decrypted by any player in the room. This includes the votes.

## Traffic

We're simulating centralized environment over decentralized transport.
Dealer acts as a server. Players only send messages to the Dealer, while Dealer publishes any changes in the room with to players.

All messages are encoded as JSON (instead of protobuf) for easier changes introduction.

There are a few message types defined.

### `State`

This is the core message of the protocol. It contains all information about current vote:
- Players list
- Issues list (with votes for each issue)
- Active issue
- Deck
- Room state (show if votes are already revealed or not)

`State` is only distributed by dealer. Moreover, this is the only message that is processed by other players. All other messages are ignored (although current encryption allows to read any message).

Dealer publishes a new `State` message on any changes in the state of the room (new player, someone's vote, adding issue).

### `PlayerVote`

Sent by any player to vote for current issue. Contains `IssueID` and `VoteValue`.

### `PlayerOnline`

Sent by all players periodically to show the dealer that players are online.

### `PlayerOffline`

Sent by any player when leaving room or closing the app to show that the user is offline.

# A note on version 2

Version 2 will address the main drawbacks of version 1.

1. Minimize pseudo-decentralization over dealer  
E.g. Online/Offline messages should not spawn a new `State` message, but be processed by all players.

2. Lower network load
Encode messages in protobuf instead of JSON.

3. Allow players to verify vote of a player
Introduce votes signatures.

4. Only send modified parts of `State`
E.g. don't duplicate players list if it did not change. 

The ideas about protocol v2 can be tracked and posted here: 
- https://github.com/six78/2-story-points-cli/issues/82
