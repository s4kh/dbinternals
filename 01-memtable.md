# Stage 1: Memtable

**Build:** an in-memory sorted store in Go, standard library only.

## Operations
- `Put(key, value)`
- `Delete(key)`
- `Get(key)`
- an ordered iterator over all entries
- a byte-size counter

## Requirements
1. Every entry carries a **sequence number** (a counter you assign) and a **tombstone flag**.
2. `Delete` stores a tombstone entry. It never removes the key.
3. `Get` must tell apart three outcomes: **found**, **deleted** (a tombstone), and **absent** (never seen). Pick a return signature that makes the difference impossible to miss.
4. The size counter follows a rule you define: key bytes + value bytes + a fixed per-entry overhead. Write the rule in a comment.
5. The iterator returns entries in key order.

## Decisions to make (justify each in a short comment)
- Which data structure, and what each alternative gives up (think about insert, point lookup, ordered iteration).
- Keep **every version** of a key, or only the newest? What does your choice do to the size counter on an overwrite?
- What does the iterator return for a tombstone? Hide it or show it? Think about what the flush will need later.

## Tests to write
- Put, then get.
- Overwrite, then get returns the latest.
- Delete, then get says deleted. Delete of a never-written key, then get says deleted too. (Is that what you want? Decide and defend it.)
- Iteration order after inserting keys in random order.
- The size counter after put, overwrite, and delete: what does each do?

## Avoid
Don't make it thread-safe yet. That comes at Stage 4, and you should feel why it is needed.

## When done
Send the file path and your confidence on each decision (low, medium, high). I'll review it and try to break it.
