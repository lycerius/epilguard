# Epilguard
Epilguard analyzes for photosensitive content in videos, specifically flashing content. The project is compliant with [ITU-R BT.1702](https://www.itu.int/dms_pubrec/itu-r/rec/bt/R-REC-BT.1702-0-200502-I!!PDF-E.pdf), which details how to identify photosensitive hazards in video content.

## Using Epilguard
You can download the latest release of epilguard for your platform from github. Before you can use Epilguard, download a build of [FFMpeg](https://www.ffmpeg.org/) and add it to your PATH.

You can now use epilguard from your terminal.
```terminal
epilguard
```

## Creating Reports
To create a hazard report for a video:
```terminal
epilguard video.mp4 report/directory
```

This will analyze video.mp4 and create the hazard report in report/directory. The report consists of 4 seperate files:

```tree
report-directory/
├── [timestamp]-[videoname]-Accumulation.csv
├── [timestamp]-[videoname]-Flashes.csv
├── [timestamp]-[videoname]-FrameFlashes.csv
└── [timestamp]-[videoname]-Report.json
```

**Artifact Descriptions**:
* **Accumulation** - How the brightness accumulated over frames before inversion.
* **Flashes** - A compressed version of accumulation that details the maximum brightness achieved over X frames before inversion
* **FrameFlashes** - Like flashes, but uses frame indexes instead of frame count
* **Report** - The descriptive hazard report which describes each hazard and where it started/ended in the video.

*Note: You can plot the CSV files using a common plotting utility (such as Excel or PyPlot) to visualize the algorithm.*

## Hazard Report
A Hazard Report is a JSON file that contains a number of hazards and descriptions that explain why they were considered hazardous.

Example Hazard Report:
```json
{
  
    "createdOn": DateString,
    "hazards": [
        {
            "start": number,
            "end": number,
            "hazardType": string
        },
        ...
    ]
}
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
