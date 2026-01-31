# BigBoxDB Server

I've decided to rewrite the BigBoxDB server completely and remove NextJS depedancy completely, and teach myself Golang in the process!

This exists purely as an API server/host for all related files. I want to separate the webfront out to be almost completely standalone to make this a bit more modular/selfhostable. Who knows what kind of boxes will show up and from what clients!

## Current Features

- `/api` routes talk to MariaDB backend
- Unzipping and processing source files successfully
- Fun relations/automatic foreign keys with GoORM
- `.env` files for seeding and config
- Users table with randomly generated API key

## Maybe Features?
- Maybe it could do the tiff->webp/glb conversion on this end instead of python? (though I'd have to handle potentially really big zip files...)