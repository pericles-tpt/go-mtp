# go-mtp
Go-mtp is a wrapper for https://github.com/libmtp/libmtp. I started this project to allow me to access the files on my Garmin Forerunner 255 in an automated way via the MTP protocol.

Although there are other projects that do something similar, i.e. `go-mtps` and `go-mtpx`, they're not in active development and either:
1. don't allow me to access the storage on my Garmin Forerunner 255
2. don't include a licence so cannot be used in my other projects

They also seem to use the `libusb` library instead of the `libmtp` library that this project wraps.

## Setup
### Fedora
To start working on this project you'll need to install the `libmtp-devel` package

## Contibributing
Contributions to this project are welcome, I'll be focusing on functionality that allows me to open, read storage, transfer files and get system information from my Garmin watch.

But, if you want to implementing other bindings I'll be happy to review your MR.

## Current State
I've setup some initial bindings and a "test" scenario in my `main()` function. It's all a bit messy at the moment but it's functional on my Fedora desktop, connecting to my Garmin watch and listing the files on it.

### TODO
- Move bindings into separate packages
- Create some tests
- Check for any missing "cleanup", "release", "free" calls
- Improve error handling