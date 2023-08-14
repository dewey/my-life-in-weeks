# My Life in Weeks

This is a small fun project inspired by KrauseFx's [howisfelix.today](https://howisfelix.today/?), which, in turn, is inspired by [Wait But Why's project](https://waitbutwhy.com/2014/05/life-weeks.html).

I decided to do it based on the EXIF data of my personal photo collection in Photos.app. The caveat there is that properly GPS tagged data is only available from the point in time when I got my first iPhone (January 2008).

## Output

Here's a work in progress screenshot, with around 50% of my images correctly tagged with the location (I ran out of API credits).

<img width="1918" alt="Screenshot 2023-08-14 at 19 07 58 png@2x" src="https://github.com/dewey/my-life-in-weeks/assets/790262/357046d1-4f7a-4e7f-a331-a6028a288a12">


## Usage

This is a two-step process, first we are scanning the Photos.app library and extracting the GPS coordinates. Then we use a service like [Positionstack.com](http://positionstack.com) to reverse geocode these into Country / City information.

All this information is stored in a local SQLite database. Running the import step another time will skip all entries which are in the database already.

While the data is importing you can already watch the progress by starting the web interface of the tool (See: "Generate report" below).

### Configuration

- `POSITIONSTACK_TOKEN`: Add your API token from http://positionstack.com
- `PHOTO_PATH`: The absolute path to your Photos.app originals directory
- `LOCATION_BACKEND` (Default: positionstack): The backend used to translate coordinates to country / city information.

### Import data

Make sure [Exiftool](https://exiftool.org) is installed and available in your $PATH.

Run `./run_develop.sh` to scan your Photos.app library.

### Generate report

Run `./run_develop_web.sh` to generate the report. You can look at it on "http://localhost:8080" by default.

## Other thoughts

- It might be more efficient to dig into the SQLite database of Photos.app directly. The schema is [not straight forward](https://simonwillison.net/2020/May/21/dogsheep-photos/) though. This could be implemented as a different importer plugin.

