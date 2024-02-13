# pastebin
Simple < 150 line pastebin in Go. Made for local labs.  
The IDs generated are two word combos, hyphen separated. The words are four or less characters for easy remembering and typing into another computer.  
Pastes are set to expire after 2 minutes.

## Usage
Create:
```bash
curl -X POST localhost:4242 -d "your paste content"
wack-knot
```

Get:
```bash
curl localhost:4242/wack-knot
your paste content
```

## TODO
- [ ] Refactor to not use gorilla/mux, light weight as possible.
- [ ] Add options for custom expire times.
- [ ] Optional shell script would be nice.
