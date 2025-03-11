.PHONY = "run/greenmail"

run/greenmail:
	@docker run -d \
  -e GREENMAIL_OPTS="-Dgreenmail.setup.test.smtp -Dgreenmail.setup.test.imap" \
  -p 3025:3025 -p 3143:3143 -p 8080:8080 \
  greenmail/standalone
