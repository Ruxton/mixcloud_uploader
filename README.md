Mixcloud CLI Uploader
====================================

Overview
--------

A go CLI application that uploads MP3's (cloudcasts)

It can:
 * Add descriptions
 * Add tags
 * Add cover art
 * Parse Virtual DJ 7 Tracklist.txt sections to approximate track changes

Binaries
---------

Compiled [binaries are available on dropbox](https://www.dropbox.com/sh/bqrajt2q74vn3jx/AABBFzI4327haGgjpHQrwKHHa) for:
  * [Windows 64bit](https://www.dropbox.com/s/4hb25ooz2p3h38d/mixcloud.exe)
  * [Linux 32bit](https://www.dropbox.com/s/p2ny4njqfm966z0/mixcloud.linux)
  * [Linux 64bit](https://www.dropbox.com/s/g6p5fg7bnn9x5o2/mixcloud.linux64)
  * [OSX 64bit](https://www.dropbox.com/s/27j48w0bete7xkt/mixcloud.osx)

Source Requirements
------------

* GoLang > 1.2.1
* The Internet

Using Mixcloud CLI Uploader From Source
--------------------------------

  1. Goto http://www.mixcloud.com/developers/ and create a new application
  1. cp build/oauth_client.env.sample build/oauth_client.env
  1. Fill out build/oauth_client.env with your Oauth Application details from step 1
  1. Run bin/build
  1. Run the built packages as below from pkg/

Using Pre-compiled Packages
---------------------------

Using compiled packages:

  `mixcloud --file <path_to_filename> --cover <path_to_cover> --tracklist <path_to_tracklist>`

Meta
----

* Code: `git clone git://github.com/ruxton/mixcloud_uploader.git`
