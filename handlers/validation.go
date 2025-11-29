package handlers

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Supported DNS record types
var supportedRecordTypes = map[string]bool{
	"A":     true,
	"AAAA":  true,
	"CNAME": true,
	"MX":    true,
	"TXT":   true,
	"NS":    true,
	"PTR":   true,
	"SRV":   true,
	"CAA":   true,
	"SOA":   true,
}

// Hostname regex - RFC 1123 compliant
var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])$`)

// FQDN regex - allows subdomains
var fqdnRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.?$`)

// Email-like format for SOA admin email (with dots instead of @)
var soaEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+\.[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Zone name validation
var zoneNameRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}\.?$`)

// ValidateZoneName validates a DNS zone name
func ValidateZoneName(name string) ValidationErrors {
	var errs ValidationErrors

	if name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "zone name is required"})
		return errs
	}

	// Remove trailing dot for validation
	cleanName := strings.TrimSuffix(name, ".")

	if len(cleanName) > 253 {
		errs = append(errs, ValidationError{Field: "name", Message: "zone name exceeds maximum length of 253 characters"})
	}

	// Check each label
	labels := strings.Split(cleanName, ".")
	if len(labels) < 2 {
		errs = append(errs, ValidationError{Field: "name", Message: "zone name must have at least two labels (e.g., example.com)"})
	}

	for _, label := range labels {
		if len(label) > 63 {
			errs = append(errs, ValidationError{Field: "name", Message: fmt.Sprintf("label '%s' exceeds maximum length of 63 characters", label)})
		}
		if len(label) == 0 {
			errs = append(errs, ValidationError{Field: "name", Message: "zone name contains empty label"})
		}
		if !hostnameRegex.MatchString(label) {
			errs = append(errs, ValidationError{Field: "name", Message: fmt.Sprintf("label '%s' contains invalid characters", label)})
		}
	}

	return errs
}

// ValidateSOA validates SOA record fields
func ValidateSOA(primaryNS, adminEmail string, refresh, retry uint16, expire uint32, minimum, ttl uint16) ValidationErrors {
	var errs ValidationErrors

	// Primary NS validation
	if primaryNS == "" {
		errs = append(errs, ValidationError{Field: "soa.primary_ns", Message: "primary nameserver is required"})
	} else if !isValidHostname(primaryNS) {
		errs = append(errs, ValidationError{Field: "soa.primary_ns", Message: "invalid primary nameserver hostname"})
	}

	// Admin email validation (SOA format: admin.example.com instead of admin@example.com)
	if adminEmail == "" {
		errs = append(errs, ValidationError{Field: "soa.admin_email", Message: "admin email is required"})
	} else if !soaEmailRegex.MatchString(adminEmail) {
		errs = append(errs, ValidationError{Field: "soa.admin_email", Message: "invalid admin email format (use dots instead of @, e.g., admin.example.com)"})
	}

	// Refresh validation (typically 1 hour to 1 day)
	if refresh < 300 {
		errs = append(errs, ValidationError{Field: "soa.refresh", Message: "refresh interval should be at least 300 seconds (5 minutes)"})
	}

	// Retry validation (typically less than refresh)
	if retry < 60 {
		errs = append(errs, ValidationError{Field: "soa.retry", Message: "retry interval should be at least 60 seconds"})
	}

	// Expire validation (typically 1 week to 4 weeks)
	if expire < 3600 {
		errs = append(errs, ValidationError{Field: "soa.expire", Message: "expire time should be at least 3600 seconds (1 hour)"})
	}

	// Minimum TTL validation
	if minimum < 60 {
		errs = append(errs, ValidationError{Field: "soa.minimum", Message: "minimum TTL should be at least 60 seconds"})
	}

	// TTL validation
	if ttl < 60 {
		errs = append(errs, ValidationError{Field: "soa.ttl", Message: "TTL should be at least 60 seconds"})
	}

	return errs
}

// ValidateRecordType validates the DNS record type
func ValidateRecordType(recordType string) ValidationErrors {
	var errs ValidationErrors

	if recordType == "" {
		errs = append(errs, ValidationError{Field: "type", Message: "record type is required"})
		return errs
	}

	upperType := strings.ToUpper(recordType)
	if !supportedRecordTypes[upperType] {
		errs = append(errs, ValidationError{
			Field:   "type",
			Message: fmt.Sprintf("unsupported record type '%s'. Supported types: A, AAAA, CNAME, MX, TXT, NS, PTR, SRV, CAA", recordType),
		})
	}

	return errs
}

// ValidateRecordHost validates the host/name field of a DNS record
func ValidateRecordHost(host string) ValidationErrors {
	var errs ValidationErrors

	if host == "" {
		errs = append(errs, ValidationError{Field: "host", Message: "host is required"})
		return errs
	}

	// Allow @ for apex records
	if host == "@" {
		return errs
	}

	// Allow wildcard
	if host == "*" || strings.HasPrefix(host, "*.") {
		host = strings.TrimPrefix(host, "*.")
		if host == "" {
			return errs
		}
	}

	// Validate hostname format
	cleanHost := strings.TrimSuffix(host, ".")
	if len(cleanHost) > 253 {
		errs = append(errs, ValidationError{Field: "host", Message: "host exceeds maximum length of 253 characters"})
	}

	labels := strings.Split(cleanHost, ".")
	for _, label := range labels {
		if len(label) > 63 {
			errs = append(errs, ValidationError{Field: "host", Message: fmt.Sprintf("label '%s' exceeds maximum length of 63 characters", label)})
		}
		if label != "" && !hostnameRegex.MatchString(label) && label != "_dmarc" && !strings.HasPrefix(label, "_") {
			errs = append(errs, ValidationError{Field: "host", Message: fmt.Sprintf("label '%s' contains invalid characters", label)})
		}
	}

	return errs
}

// ValidateRecordContent validates the content based on record type
func ValidateRecordContent(recordType, content string) ValidationErrors {
	var errs ValidationErrors

	if content == "" {
		errs = append(errs, ValidationError{Field: "content", Message: "content is required"})
		return errs
	}

	upperType := strings.ToUpper(recordType)

	switch upperType {
	case "A":
		errs = append(errs, validateIPv4(content)...)
	case "AAAA":
		errs = append(errs, validateIPv6(content)...)
	case "CNAME", "NS", "PTR":
		errs = append(errs, validateHostname(content)...)
	case "MX":
		errs = append(errs, validateMX(content)...)
	case "TXT":
		errs = append(errs, validateTXT(content)...)
	case "SRV":
		errs = append(errs, validateSRV(content)...)
	case "CAA":
		errs = append(errs, validateCAA(content)...)
	}

	return errs
}

// validateIPv4 validates an IPv4 address
func validateIPv4(content string) ValidationErrors {
	var errs ValidationErrors

	ip := net.ParseIP(content)
	if ip == nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid IP address", content)})
		return errs
	}

	// Check if it's IPv4
	if ip.To4() == nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid IPv4 address (use AAAA record for IPv6)", content)})
	}

	return errs
}

// validateIPv6 validates an IPv6 address
func validateIPv6(content string) ValidationErrors {
	var errs ValidationErrors

	ip := net.ParseIP(content)
	if ip == nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid IP address", content)})
		return errs
	}

	// Check if it's IPv6 (not IPv4)
	if ip.To4() != nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is an IPv4 address (use A record for IPv4)", content)})
	}

	return errs
}

// validateHostname validates a hostname for CNAME, NS, PTR records
func validateHostname(content string) ValidationErrors {
	var errs ValidationErrors

	if !isValidHostname(content) {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid hostname", content)})
	}

	return errs
}

// validateMX validates MX record content (priority hostname)
func validateMX(content string) ValidationErrors {
	var errs ValidationErrors

	parts := strings.Fields(content)
	if len(parts) == 1 {
		// Just hostname, no priority - that's okay, priority might be handled separately
		if !isValidHostname(parts[0]) {
			errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid mail server hostname", parts[0])})
		}
	} else if len(parts) == 2 {
		// Priority and hostname
		if _, err := fmt.Sscanf(parts[0], "%d", new(int)); err != nil {
			errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid MX priority", parts[0])})
		}
		if !isValidHostname(parts[1]) {
			errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid mail server hostname", parts[1])})
		}
	} else {
		errs = append(errs, ValidationError{Field: "content", Message: "MX record should be in format 'priority hostname' or just 'hostname'"})
	}

	return errs
}

// validateTXT validates TXT record content
func validateTXT(content string) ValidationErrors {
	var errs ValidationErrors

	// TXT records can contain almost anything, but check length
	if len(content) > 65535 {
		errs = append(errs, ValidationError{Field: "content", Message: "TXT record content exceeds maximum length"})
	}

	return errs
}

// validateSRV validates SRV record content (priority weight port target)
func validateSRV(content string) ValidationErrors {
	var errs ValidationErrors

	parts := strings.Fields(content)
	if len(parts) != 4 {
		errs = append(errs, ValidationError{Field: "content", Message: "SRV record should be in format 'priority weight port target'"})
		return errs
	}

	// Validate priority
	if _, err := fmt.Sscanf(parts[0], "%d", new(int)); err != nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid SRV priority", parts[0])})
	}

	// Validate weight
	if _, err := fmt.Sscanf(parts[1], "%d", new(int)); err != nil {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid SRV weight", parts[1])})
	}

	// Validate port
	var port int
	if _, err := fmt.Sscanf(parts[2], "%d", &port); err != nil || port < 0 || port > 65535 {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid port number", parts[2])})
	}

	// Validate target
	if parts[3] != "." && !isValidHostname(parts[3]) {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid SRV target hostname", parts[3])})
	}

	return errs
}

// validateCAA validates CAA record content
func validateCAA(content string) ValidationErrors {
	var errs ValidationErrors

	parts := strings.Fields(content)
	if len(parts) < 3 {
		errs = append(errs, ValidationError{Field: "content", Message: "CAA record should be in format 'flags tag value'"})
		return errs
	}

	// Validate flags (0 or 128)
	var flags int
	if _, err := fmt.Sscanf(parts[0], "%d", &flags); err != nil || (flags != 0 && flags != 128) {
		errs = append(errs, ValidationError{Field: "content", Message: "CAA flags must be 0 or 128"})
	}

	// Validate tag
	validTags := map[string]bool{"issue": true, "issuewild": true, "iodef": true}
	if !validTags[strings.ToLower(parts[1])] {
		errs = append(errs, ValidationError{Field: "content", Message: fmt.Sprintf("'%s' is not a valid CAA tag (use issue, issuewild, or iodef)", parts[1])})
	}

	return errs
}

// ValidateRecordTTL validates the TTL value
func ValidateRecordTTL(ttl uint16) ValidationErrors {
	var errs ValidationErrors

	if ttl > 0 && ttl < 60 {
		errs = append(errs, ValidationError{Field: "ttl", Message: "TTL should be at least 60 seconds"})
	}

	return errs
}

// ValidateAddPTR validates if add_ptr is appropriate for the record type
func ValidateAddPTR(recordType string, addPTR bool) ValidationErrors {
	var errs ValidationErrors

	if addPTR {
		upperType := strings.ToUpper(recordType)
		if upperType != "A" && upperType != "AAAA" {
			errs = append(errs, ValidationError{Field: "add_ptr", Message: "add_ptr can only be set for A or AAAA records"})
		}
	}

	return errs
}

// ValidateRecord validates all fields of a DNS record
func ValidateRecord(recordType, host, content string, ttl uint16, addPTR bool) ValidationErrors {
	var errs ValidationErrors

	errs = append(errs, ValidateRecordType(recordType)...)
	errs = append(errs, ValidateRecordHost(host)...)
	errs = append(errs, ValidateRecordContent(recordType, content)...)
	errs = append(errs, ValidateRecordTTL(ttl)...)
	errs = append(errs, ValidateAddPTR(recordType, addPTR)...)

	return errs
}

// isValidHostname checks if a string is a valid hostname
func isValidHostname(hostname string) bool {
	if hostname == "" {
		return false
	}

	// Remove trailing dot
	hostname = strings.TrimSuffix(hostname, ".")

	if len(hostname) > 253 {
		return false
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		// Allow underscore for SRV and other special records
		if !hostnameRegex.MatchString(label) && !strings.HasPrefix(label, "_") {
			return false
		}
	}

	return true
}
