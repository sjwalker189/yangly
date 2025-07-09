# Yangly

> This project is a work in progress

**Yangly** is a cli tool for generating TypeScript definitions from YANG modules.


## Installation
```bash
npm install --save-dev --save-exact yangly
```

## Usage

```bash
# Generate types
npx yangly gen --path yangs --out types
```

## TODO:
- [x] Generate types for YANG schemas
- [ ] Generate types for YANG RPC
- [ ] Generate types for YANG actions
- [ ] Generate types for YANG notifications
- [ ] Format type output
- [ ] Generate runtime validators
- [ ] Multiplatform support (only linux is supported currently) 

