#Go Downloading
##Requirements
- Check if range downloads are supported, if so do a parallel download.
- Handle errors and retries
- Check download file for integrity
- Bonus point if you provide benchmarks for various chunk sizes for parallel downloads

##Usage
The program allows for a user to specify both the url to download from and how many threads to split the download up into (if supported).  The program will automatically calculate chunk sizes (in bytes) based on the number of threads specified.  The default is 100 threads because it seemed to work the best in my testing.
###Arguments
| Name    | Usage     | Default Value |
| -----   | -----     | ------------- |
| threads | -threads=5  | 100           |
| url     | -url="www.google.com"      | http://storage.googleapis.com/vimeo-test/work-at-vimeo.mp4 |

##Benchmarks
TODO

##Contributors
* **Gian Biondi** <gianbiondijr@gmail.com>
