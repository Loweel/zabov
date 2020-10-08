# zabov

Tiny replacement for piHole DNS filter

Still Work in progress, usable.

Idea is to produce a very simple, no-web-interface , IP DNS blocker.

# INSTALL

Zabov requires golang 1.13 or later.

<pre>
git clone https://git.keinpfusch.net/Loweel/zabov.git
cd zabov
go get
go build -mod=vendor
</pre>

Then, edit config.json: please notice config.json must be in the same folder of the executable you run.


Just a few words about "singlefilters" and "doublefilters":

Data must be downloaded from URLs of blacklist mantainers.They may come in different formats.

There are two kinds of blacklists:

One is the format zabov calls "singlefilter", where we find a single column , full of domains:

<pre>
domain1.com
domain2.com
domain3.com
</pre>

The second is the format zabov calls "doublefilter" (a file in "/etc/hosts" format, to be precise), where there is an IP, usually localhost or 0.0.0.0 and then the domain:

<pre>
127.0.0.1 domain1.com
127.0.0.1 domain2.com
127.0.0.1 domain3.com
</pre>

This is why configuration file has two separated items.

The config file should look like:

<pre>
{
    "zabov": {  
        "port":"53", 
        "proto":"udp", 
        "ipaddr":"127.0.0.1",
        "upstream":"./dns-upstream.txt",
        "cachettl": "4",
        "killfilettl": "12",
        "singlefilters":"./urls-hosts.txt" ,
        "doublefilters":"./urls-domains.txt", 
        "blackholeip":"127.0.0.1",
        "hostsfile":"./urls-local.txt"
    }

}



</pre>

Where:

- port is the port number. Usually is 53, you can change for docker, if you like
- proto is the protocol. Choices are "udp", "tcp", "tcp/udp"
- ipaddr is the port to listen to. Maybe empty, (which will result in listening to 0.0.0.0) to avoid issues with docker.
- upstream: file containing all DNS we want to query :  each line in format IP:PORT
- cachettl: amount of time the cache is kept (in hours)
- killfilettl: refresh time for _killfiles_
- singlefilters: name of the file  for blacklists following the "singlefilter" schema.(one URL per line)
- doublefilters: name of the file, for blacklists following the "doublefilter" schema.(one URL per line)
- blackholeip: IP address to return when the IP is banned. This is because you may want to avoid MX issues, mail loops on localhost, or you have a web server running on localhost
- hostsfile: path where you keep your local blacklistfile : this is in the format "singlefilter", meaning one domain per line, unlike hosts file.

# DOCKER
Multistage Dockerfiles are provided for AMD64, ARMv7, ARM64V8

# TODO:

- ~~caching~~
- monitoring port


