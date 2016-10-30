# Oolong
## What is this?
Oolong is a client for [wirelesstag.net](http://www.wirelesstag.net) designed to
run in the background and periodically poll the wirelesstag API and fetch new
data.

## Why?
I don't like having the only copy of my data stored in the cloud and the graph
viewing on their site is too simple for what I want (although it is probably
fine for most people).  I already use OpenTSDB+Grafana for other personal
projects, so I wanted to reuse that for viewing data collected by the tags as
well.

## How to Run
1.  First you'll need an OAuth client id+secret from their website.  You can get
    one here: [https://mytaglist.com/eth/oauth2_apps.html](https://mytaglist.com/eth/oauth2_apps.html)
2.  Copy the example config, fill out your oauth and opentsdb details.
3.  Initialize the client: `$ ./oolong init`
    -  The client will start an HTTP server and display the link to go to in
    your browser.
    -  Approve the access on the wirelesstag/mytaglist website.  The page should
    display the name you gave the OAuth app in step 1.
    -  Check the output of the client, it should display `Got access token: xxx...`
    and then exit.
4.  Run the client: `$ ./oolong run`
    -  That's it.  The client will run until something fails or you kill it.
