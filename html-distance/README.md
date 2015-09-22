# html-distance

html-distance is a go library for computing the proximity of the HTML pages. The implementation similiarity fingerprint is Charikar's simhash. 

We used BK Tree (Burkhard and Keller) for verifying if a fingerprint is closed to a set of fingerprint within a defined proximity distance. 

Distance is the hamming distance of the fingerprints. Since fingerprint is of size 64 (inherited from hash/fnv), Similiarity is defined as 1 - d / 64.

In normal scenario, similarity > 95% (i.e. d>3) could be considered as duplicated html pages.


## Get the source

```
go get github.com/yahoo/gryffin/html-distance/...
```

## Install 

```
go install github.com/yahoo/gryffin/html-distance/cmd/html-distance
```

## Command Line Interface

```
Usage of html-distance:

    html-distance url1 url2
```

Example 1
```
$ html-distance https://www.flickr.com/photos/120759744@N07/20389369791/ https://www.flickr.com/photos/120759744@N07/20374523532/in/photostream/

Fetching https://www.flickr.com/photos/120759744@N07/20389369791/, Got 200
Fetching https://www.flickr.com/photos/120759744@N07/20374523532/in/photostream/, Got 200
Feature distance is 0. HTML Similarity is 100.00%
```

Example 2
```
$ html-distance https://www.yahoo.com/politics/kasichs-reception-on-gay-marriage-important-126109300441.html https://www.yahoo.com/tech/s/verizon-drop-phone-contracts-end-discounted-phones-201530971--finance.html

Fetching https://www.yahoo.com/politics/kasichs-reception-on-gay-marriage-important-126109300441.html, Got 200
Fetching https://www.yahoo.com/tech/s/verizon-drop-phone-contracts-end-discounted-phones-201530971--finance.html, Got 200
Feature distance is 2. HTML Similarity is 96.88%
```

Example 3
```
$ html-distance https://www.flickr.com/photos/120759744@N07/20389369791/ https://www.yahoo.com/tech/s/verizon-drop-phone-contracts-end-discounted-phones-201530971--finance.html

Fetching https://www.flickr.com/photos/120759744@N07/20389369791/, Got 200
Fetching https://www.yahoo.com/tech/s/verizon-drop-phone-contracts-end-discounted-phones-201530971--finance.html, Got 200
Feature distance is 9. HTML Similarity is 85.94%
```
