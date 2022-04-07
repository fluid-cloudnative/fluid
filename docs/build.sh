#!/bin/bash

langs="en zh"
for lang in $langs
do
  cp ${lang}/dev/api_doc.md ${lang}/dev/api_doc.html
  python3 scripts/mergeByTOC.py ${lang}/
done

./scripts/genDoc.sh

for lang in $langs
do
  xvfb-run wkhtmltopdf ./en/dev/api_doc.html api.pdf
  python3 scripts/mergePDF.py ${lang}
done

for lang in $langs
do
  rm ${lang}/dev/api_doc.html
  rm ${lang}/doc.md
  rm api.pdf
  rm output_${lang}.pdf
done
