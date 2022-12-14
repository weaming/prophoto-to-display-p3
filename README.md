# ProPhoto to Display P3

Convert from ProPhoto to Display P3.

ProPhoto is exported from RAW file, and Display P3 is used widely around the world.

## Usage

```
mkdir $HOME/icc
curl https://raw.githubusercontent.com/weaming/prophoto-to-display-p3/main/DisplayP3.icc > ~/icc/DisplayP3.icc
go install github.com/weaming/prophoto-to-display-p3@latest
```

## [Embed icc into output jpeg](https://exiftool.org/forum/index.php?topic=1596.0)

First, you need a valid ICC profile to write into the file. You can extract it from any other image containing
a profile with a command like this:

    exiftool -icc_profile -b src.jpg > profile.icc

Then this command writes the profile to "dest.jpg":

    exiftool "-icc_profile<=profile.icc" dest.jpg
