# kruise-api

Schema of the API types that are served by Kruise.

## Purpose

This library is the canonical location of the Kruise API definition.

We recommend using the go types in this repo. You may serialize them directly to JSON.

## Compatibility matrix

| Kubernetes Version in your Project | Import Kruise-api < v0.10  | Import Kruise-api >= v0.10 |
|------------------------------------|----------------------------|----------------------------|
| < 1.18                             | v0.x.y (x <= 9)            | v0.x.y-legacy (x >= 10)    |
| >= 1.18                            | v0.x.y-1.18 (7 <= x <= 9)  | v0.x.y (x >= 10)           |

## Where does it come from?

`kruise-api` is synced from [https://github.com/openkruise/kruise/tree/master/apis](https://github.com/openkruise/kruise/tree/master/apis).
Code changes are made in that location, merged into `openkruise/kruise` and later synced here.

## Things you should NOT do

[https://github.com/openkruise/kruise/tree/master/apis](https://github.com/openkruise/kruise/tree/master/apis) is synced to here.
All changes must be made in the former. The latter is read-only.

