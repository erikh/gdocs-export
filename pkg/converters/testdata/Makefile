generate:
	for dir in $$(find . -maxdepth 1 -type d | tail -n +2); \
	do \
		dir=$$(basename $$dir); \
		cd $$dir; \
		for format in md html; \
		do \
			if [ -d assets ]; then \
				go run ../../../../cmd/gdexport c -a assets $$format $$dir.json > $$dir.$$format; \
			else \
				go run ../../../../cmd/gdexport c $$format $$dir.json > $$dir.$$format; \
			fi; \
		done; \
		cd ..; \
	done
