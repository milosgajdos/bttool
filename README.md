**!!!! THIS IS HIGHLY EXPERIMENTAL, INCOMPLETE AND PROBABLY QUITE BUGGY TOOL AT THE MOMENT !!!!**

# bttool
BitTorrent metainfo tool

```bttool``` allows you to generate and decode BitTorrent metainfo files. 

Currently the tool implements two subcommands: ```encode``` and ```decode```

## encode
Encode subcommand reads in [YAML](http://en.wikipedia.org/wiki/YAML) encoded manifest file and generates bencoded BitTorrent metainfo file as per the specified configuration. By default it will dump a stream of bencoded bytes into ```stdout```. 

The following commands should run without any problems:
- ```bttool encode manifest.yml -outfile some.torrent``` - reads in ```manifest.yml``` manifest and generates ```some.torrent``` metainfo file
- ```bttool encode manifest.yml``` - reads in ```manifest.yml``` manifest and dumps bencoded byte stream into ```stdout```

#### manifest
Following is an example of suggested manifest format:

```yaml
trackers:
    - udp://tracker.openbittorrent.com:80
    - udp://tracker.openbittorrent.com:80
data:
  src: /some/src/path
  dst: /some/random/path
piecelength: 32KiB
author: Johnny Bravo
comment: This is just a test
```

## decode
Decode subcommand by default reads bencoded files from ```stdin``` and dumps the decoded data into ```stdout```. You can specify a format into which the decoded data should be encoded. Currently only ```json``` data is supported.

The following commands should run without any problems:
- ```cat test.torrent | bttool decode  -format json -outfile test.json```
    - decodes ```test.torrent``` into ```json``` and dump the output into ```test.json``` file
- ```cat test.torrent | bttool decode``` 
    - decodes ```test.torrent``` into text, dumps the content to ```stdout```
- ```bttool decode test.torrent -format json -outfile test.dump``` 
    - decodes ```test.torrent``` into ```json```, dumps the content into ```test.dump```

## validate
Validate subcommand by default reads bencoded metainfo files from ```stdin``` and validates it against data passed it via ```-data``` command line parameter. If the validation fails, the tool exits with non-zero return code. You can use ```-verbose``` command line flag which will print out validation result for each metainfo file piece.

The following commands should run withouth any problems:
- ```cat test.torrent| bttool validate -data=/path/to/data```
    - validates ```test.torrent``` against ```/path/to/data```
- ```bttool validate -data=data=/path/to/data test.torrent```
    - same as above, but torrent data is passed as a command line argument
- ```bttool validate -verbose -data=data=/path/to/data test.torrent```
    - same as above, but run in verbose mode

## Usage
Get the package:

```
go get -vu
```

or clone the repo

```
git clone https://github.com/milosgajdos83/bttool.git
```

Once the initial clone is done you can get the updated version from master via ```go get -vu```

Build:

```
go build
```

## TODO
- tests
- ```send``` commands
