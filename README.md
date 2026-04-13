# go-mtp
Go-mtp is a wrapper for https://github.com/libmtp/libmtp.

Although there are other projects that do something similar, i.e. `go-mtps` and `go-mtpx`, they're not in active development and either:
1. fail when I try to access the storage on my Garmin watch
2. don't include a licence so cannot be used in my other projects

They also seem to use the `libusb` library instead of the `libmtp` library that this project uses.

## Setup
### Fedora
To start working on this project you'll need to install the `libmtp-devel` package

## Contibributing
Contributions to this project are welcome, I'll be focusing on functionality that allows me to open, read storage, transfer files and get system information from my Garmin watch.

But, if you want to implement other bindings I'll be happy to review your MR.

### TODO
- Move bindings into separate packages [WON'T DO]
- Create some tests
- Check for any missing "cleanup", "release", "free" calls [DONE]
- Improve error handling [DONE]
