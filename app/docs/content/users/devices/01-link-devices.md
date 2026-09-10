---
title: Link Devices
section: devices
order: 1
summary: Connect another browser or desktop app to the same account and Spaces.
---

Pairing signs another browser or desktop app in to an existing account. Each
client gets its own Session, which you can remove later. Both clients confirm
matching emoji before account access changes.

## Link with a code

1. Open the account you want to share and choose **Link My Device** from setup
   or settings.
2. Choose **Generate code for another device**.
3. Open Spacewave in another browser or the desktop app. From Home, choose
   **Link My Device** and enter the 8-character code. You can also follow the
   pairing link.
4. Compare the emoji on both clients and confirm that they match.

If the emoji differ, reject the connection and start again. A code expires
and can be replaced from the client that generated it.

## When both clients already use accounts

Entering a code from within an account offers four outcomes: sign in to either
account on the other client, or merge into either account. Both clients show
which accounts and machines are involved before confirmation.

Signing in keeps the two accounts separate. Merging moves the source account's
Spaces and Sessions into the selected destination. The destination keeps its
account identity, settings, and local or cloud storage provider. Existing Space
IDs and external sharing permissions remain intact.

A permission or Session-capacity problem blocks the merge and identifies what
needs attention. Source data remains available for recovery and retry. Other
Sessions follow the account's verified transition when they reconnect. Reconnect
Sessions from an earlier local merge before moving that destination account
again.

## Link with direct signaling

Choose **Show QR code** on the first client. On the other client, scan the code
or open its pairing link, then send its answer back to the first client. Compare
and confirm the emoji on both clients.

Direct signaling is available for local and cloud accounts. It establishes the
peer connection directly; cloud accounts still contact their provider to
authorize the new Session.

## After connecting

**Account connected** means the new Session has durable account access. For a
local account, Spacewave copies each Space's files in the background. Keep a
client with the data online until copying completes. Copy status reports
progress and any interruption, and resumes when the source is reachable again.
After the copy completes, the receiving client can read that data independently.

Changes and new Spaces continue to synchronize through the account. Removing a
Session stops its future access and synchronization; it does not erase data
that client already retained.

You can pair two browsers without installing anything. The desktop download
page offers the published macOS, Windows, and Linux builds with platform-specific
instructions.

Account pairing is separate from adding a managed Device to a Space. Use the
managed Device flow when you want to connect a machine or agent for Spacewave to
operate.
