# BigBoxDB Server

I've decided to rewrite the BigBoxDB server completely and remove NextJS depedancy completely, and teach myself Golang in the process!

This exists purely as an API server/host for all related files. I want to separate the webfront out to be almost completely standalone to make this a bit more modular/selfhostable. Who knows what kind of boxes will show up and from what clients!

## Current Features

- `/api` routes talk to MariaDB backend
- Unzipping and processing source files successfully
- Fun relations/automatic foreign keys with GORM
- `.env` files for seeding and config
- Users table with randomly generated API key
- tif/webp conversion to 3d model through vips/gltf magic

## Notes on image names

Besides the box_type, I also check file names for which face the texture/box should go on. All games will have these for example (either webp or tif):

- back
- bottom
- front
- left
- right
- top

Boxes with gatefolds (usually at the front) will add these:

- gatefold_right
- gatefold_left

Boxes with gatefolds at the front AND back (I'm looking at you Black & White!) instead looks like this

- gatefold_front_left
- gatefold_front_right
- gatefold_back_left
- gatefold_back_right

Boxes with double gatefolds in the front, look like this

- gatefold_front_left — left flap outer face
- gatefold_front_left_back — left flap inner face (what you see when it swings open)
- gatefold_front_right — right flap outer face
- gatefold_front_right_back — right flap inner face