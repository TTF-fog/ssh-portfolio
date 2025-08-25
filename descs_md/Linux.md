Erm, i'd like to interject there for a second.
====
<mdimg  "images/img_4.png-60-60" >

I've daily driven Arch (btw) Linux for over 2 years, as well as developed CI/CD scripts meant to run on Linux Systems.
I've also used Raspbian and Debian for RoboCup robots.

Projects (Scripts)-
- [Flarial](https://flarial.xyz)
  created a series of Github Actions meant to streamline development. These were
  - [DLL](https://github.com/flarialmc/dll/blob/main/.github/workflows/build-latest.yml)
   Automatically pushes the DLL to the respective CDN Path depending on commit specifiers
  - [Launcher](https://github.com/flarialmc/launcher/blob/main/.github/workflows/autoupdater.yml)
    Acts as an auto-updater, writing a bumped version to file metadata, and pushing updated version and latest version number to CDN. Client code takes care of checking for updates
  - [CDN](https://github.com/flarialmc/newcdn/blob/main/.github/workflows/dllhash.yml)  
   collected SHA hashes for the latest dll, allowing the launcher to check if the downloaded version was the same. this saved bandwidth for users, by preventing redownloads every launch