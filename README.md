

# twitcastd

download video from twitcasting

# usage


for example download this [video](https://twitcasting.tv/ogurayui1017/movie/702488554)

`twitcastd -u "https://twitcasting.tv/ogurayui1017/movie/702488554" -o output.mp4`

If you need to use a proxy, please set the environment variable `HTTPS_PROXY`.

# other

The loading speed of the downloaded video will be slow.
It is recommended to use ffmpeg to move the location of the metadata to the front.

for example:

`ffmpeg -i <input> -c copy -movflags faststart <output>`

# ref

https://github.com/niseyami/twitcasting-dl