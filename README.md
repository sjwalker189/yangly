# yangly

> This project is a work in progress

**yangly** is a tool to produce TypeScript definitions from YANG modules.


## Requirements

- [Go](https://go.dev/) version 1.24 or later


## Example Usage

```bash
# Build
go build main.go

# Generate yang schema types
./yangly -p path/to/yangs -o dist/types
```


## TODO
I might not do all of these:
- [ ] Emit doc blocks including descriptions, default values, etc
- [x] Generate Schema definitions
- [ ] Generate Schema validators (e.g. zod mini)
- [ ] Generate RPC definitions
- [ ] Generate Alarm definitions
- [ ] Generate Notification definitions
- [ ] Formatting TypeScript output
- [ ] Resolve leafref types
- [ ] Structured (pretty) logging with log levels
- [ ] Cross-compile binaries for other platforms (currently linux x64 only)

