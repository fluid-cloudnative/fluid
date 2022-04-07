api:
	cp zh/dev/api_doc.md zh/dev/api_doc.html;
merge:
	python3 scripts/mergeByTOC.py zh/
	python3 scripts/mergeByTOC.py en/
md2pdf:merge
	./scripts/genDoc.sh
api2pdf:api
	xvfb-run wkhtmltopdf zh/dev/api_doc.html api.pdf
build:md2pdf api2pdf
	python3 scripts/mergePDF.py zh;
	python3 scripts/mergePDF.py en
clean:
	rm api.pdf;
	rm output*.pdf;
	rm en/doc.md;rm zh/doc.md;
	rm zh/dev/api_doc.html;
	