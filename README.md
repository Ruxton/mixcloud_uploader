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

Compiled [binaries are available on dropbox](https://www.dropbox.com/sh/492dk6xvkywa1m6/AABn-CxxVjn4_B5bt6r6WqY8a/latest) for:
  * [Windows 64bit](https://www.dropbox.com/sh/492dk6xvkywa1m6/AADyNRcqIs3XZhVrnpp-xBUva/latest/mixcloud.exe)
  * [Linux 32bit](https://www.dropbox.com/sh/492dk6xvkywa1m6/AAAWQUhp0VtISWvifR69XmBEa/latest/mixcloud.linux)
  * [Linux 64bit](https://www.dropbox.com/sh/492dk6xvkywa1m6/AAAWQUhp0VtISWvifR69XmBEa/latest/mixcloud.linux64)
  * [OSX 64bit](https://www.dropbox.com/sh/492dk6xvkywa1m6/AAAWQUhp0VtISWvifR69XmBEa/latest/mixcloud.osx)

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
