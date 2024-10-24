import PyPDF2
import sys

lang=sys.argv[1]
offset=0
merger=PyPDF2.PdfFileMerger()
target=[]
target.append("output_{}.pdf".format(lang))
target.append("api.pdf")
output="docs_{}.pdf".format(lang)

for pdf in target:
    merger.merge(offset,pdf)
    pn=PyPDF2.PdfFileReader(pdf).getNumPages()
    offset+=pn
merger.write(output)
