# My Personal Browser Home Page


### TODO list and future ideas

- automated testing and deployment pipeline
- canary deployment (show version on web page so we can test the difference)
- rate limiting
- api composition
- tracing

- two way TLS authentication 
	- allows me to distinguish my connections from others
	- site should work even if mutual tls fails, but if it succeeds the service will know that it is me connecting not someone else

- modular feeds
	- instead of having a generic rss parsing logic, make it modular
	- instead of adding rss urls, add rss objects
	- rss objects comply with the same interfaces, but internally they can parse feeds differently
		(eg, for Hacker News, it can use the comment instead of the url tag for the RSS url, so links direct to the comments not the actual site)
	- there can be a generic rss object, when no special functionality is needed
	- option to have aggregated feed built from multiple sources (can be great, when dont want to dedicate a separate section to some feeds, eg. rarely updating feeds)

- metrics
	- dashboard and feed level
		- request duration
		- request frequency
		- request status
		- errors
		- timeout count and duration
		- memory, cpu and network usage


- rust, podman, buildah (fly.io?)

- new GUI?
	- show EUR/HUF/USD 
	- show NOK shareprice
	- show countdown widget
	- show scene releases
	- package tracking widget (is there a public api or something like that  we can use, to add tracking for custom tracking numbers)


- FFXIV daily reset / weekly reset countdown