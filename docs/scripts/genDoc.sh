#!/bin/bash


MAINFONT="WenQuanYi Micro Hei"
MONOFONT="WenQuanYi Micro Hei Mono"

# MAINFONT="Tsentsiu Sans HG"
# MONOFONT="Tsentsiu Sans Console HG"

#_version_tag="$(date '+%Y%m%d').$(git rev-parse --short HEAD)"
_version_tag="$(date '+%Y%m%d')"

# default version: `pandoc --latex-engine=xelatex doc.md -s -o output2.pdf`
# used to debug template setting error
lang="en zh"

for d in ${lang}
do
  if [ $d = "en" ]; then
      docs_title=" Fluid Documentation"
  else
      docs_title=" Fluid 用户文档"
  fi
  pandoc -N --toc --smart --latex-engine=xelatex \
  --template=templates/template.tex \
  --columns=120 \
  --listings \
  -V title="$docs_title" \
  -V author="Fluid" \
  -V date="${_version_tag}" \
  -V CJKmainfont="${MAINFONT}" \
  -V fontsize=12pt \
  -V geometry:margin=1in \
  "$d/doc.md" -s -o "output_$d.pdf"
done
