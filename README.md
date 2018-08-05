# Epilguard
Epilguard analyzes for photosensitive content in videos. The project is compliant with [ITU-R BT.1702](https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.1702-0-200502-I!!PDF-E.pdf), which details how to identify photosensitive hazards in video content.

## Using Epilguard
You can download the latest release of epilguard for your platform from github. Before you can use Epilguard, download a build of [FFMpeg](https://www.ffmpeg.org/) and add it to your PATH.

You can now use epilguard from your terminal
```terminal
epilguard
```

## Building from Source

### Installing GOLANG
Install the latest build of golang for your platform.
1. Download Go [here](https://golang.org/dl/)
2. Follow this [guide](https://golang.org/doc/install)


### Cloning and Building Epilguard
Clone epilguard using go get
```terminal
go get -u github.com/lycerius/epilguard
```
This will place epilguard source in your $GOPATH, and build it for you.
