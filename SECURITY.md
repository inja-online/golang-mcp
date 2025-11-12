# Security Policy

## Supported Versions

We actively support and provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.0.1   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do NOT create a public GitHub issue

Security vulnerabilities should be reported privately to prevent potential exploitation.

### 2. Report via Email

Please email security concerns to: **security@inja-online.com** (or the project maintainer's email if different)

Include the following information:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)
- Your contact information (optional, for follow-up questions)

### 3. What to Expect

- **Acknowledgment**: You will receive an acknowledgment within 48 hours
- **Initial Assessment**: We will assess the vulnerability within 7 days
- **Updates**: We will provide updates on the status of the vulnerability
- **Resolution**: We will work to resolve the issue and release a fix

### 4. Disclosure Policy

- We will coordinate with you on the disclosure timeline
- We will credit you in the security advisory (unless you prefer to remain anonymous)
- We will not disclose your identity without your permission

## Security Best Practices

When using MCP Go Server:

1. **Keep dependencies updated**: Regularly update Go modules and dependencies
2. **Review permissions**: Be cautious with the `DISABLE_NOTIFICATIONS` environment variable
3. **Validate inputs**: The server validates commands, but always validate inputs in your integrations
4. **Use secure channels**: When integrating with MCP clients, use secure communication channels
5. **Monitor logs**: Review server logs for suspicious activity

## Known Security Considerations

### Command Execution

MCP Go Server executes Go commands on behalf of users. The server includes validation to prevent command injection, but:

- Commands run with the same permissions as the server process
- Always run the server with appropriate user permissions
- Review and audit command execution in production environments

### Permission System

The server includes a permission system that can prompt users before executing commands. Consider:

- Using `DISABLE_NOTIFICATIONS=true` only in trusted environments
- Implementing additional authorization layers for production use
- Reviewing permission prompts carefully

### Network Access

Some tools may make network requests (e.g., fetching package documentation):

- Network requests are made to trusted sources (go.dev, proxy.golang.org)
- Review network access in restricted environments
- Consider using a proxy for additional control

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 0.0.1 → 0.0.2)
- Documented in the [CHANGELOG.md](CHANGELOG.md)
- Announced via GitHub releases
- Tagged with security labels

## Security Checklist for Contributors

When contributing code:

- [ ] Review code for potential security issues
- [ ] Validate all user inputs
- [ ] Avoid command injection vulnerabilities
- [ ] Use secure defaults
- [ ] Document security considerations
- [ ] Update this file if security-related changes are made

## Additional Resources

- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Vulnerability Database](https://pkg.go.dev/vuln)

Thank you for helping keep MCP Go Server secure!
