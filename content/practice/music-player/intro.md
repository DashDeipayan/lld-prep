**The prompt:** *"Design a music player like Spotify's. Focus on the player, not the streaming backend."*

This problem tests **behaviour over time**, not just a class list. Strong candidates do three things well:

1. They model the player as a **state machine** (State pattern) instead of scattering `if (isPlaying)` checks.
2. They get the **queue semantics** right: shuffle that doesn't lose your place, three repeat modes, and "previous" that restarts the song.
3. They spot the **race** between "track ended" from the audio thread and "Next" from the user.

| Phase | Interview time |
|---|---|
| Clarify requirements | 3–5 min |
| Entities, class diagram | 8 min |
| State machine | 5 min |
| Queue algorithms | 7 min |
| Extensibility and concurrency | 5 min |
| Code | 15 min |
| Follow-ups | 5 min |
