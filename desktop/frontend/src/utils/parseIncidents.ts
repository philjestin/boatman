import type { Message, Incident } from '../types';

// Normalize severity strings like "MEDIUM-HIGH", "P1", "Critical" into our severity type
function normalizeSeverity(raw: string): Incident['severity'] {
  const s = raw.toLowerCase().trim();
  if (/urgent|critical|p[01]|sev[- ]?[01]/i.test(s)) return 'urgent';
  if (/high|p2|sev[- ]?2|medium[- ]high/i.test(s)) return 'high';
  if (/medium|moderate|p3|sev[- ]?3/i.test(s)) return 'medium';
  if (/low|minor|p4|sev[- ]?4/i.test(s)) return 'low';
  // Default for unrecognized
  return 'medium';
}

// Detect source from message content
function parseSource(text: string): Incident['source'] {
  if (/linear|EMP-\d+|ticket/i.test(text)) return 'linear';
  if (/bugsnag|exception|crash/i.test(text)) return 'bugsnag';
  if (/datadog|monitor|metric|alert/i.test(text)) return 'datadog';
  if (/slack|channel|thread/i.test(text)) return 'slack';
  return 'datadog';
}

// Parse status from message content
function parseStatus(text: string): Incident['status'] {
  if (/✅\s*(Fixed|Resolved)/i.test(text) || /status:\s*resolved/i.test(text)) return 'resolved';
  if (/❌\s*(Failed|Broken)/i.test(text)) return 'failed';
  if (/🔧\s*Fixing|applying fix|deploying fix/i.test(text) || /status:\s*fixing/i.test(text)) return 'fixing';
  if (/🧪\s*Testing|running tests/i.test(text) || /status:\s*testing/i.test(text)) return 'testing';
  if (/investigating|investigation|root cause/i.test(text)) return 'investigating';
  return 'new';
}

// Extract Linear ID from text
function extractLinearId(text: string): string | undefined {
  const match = text.match(/\b([A-Z]+-\d+)\b/);
  return match ? match[1] : undefined;
}

// Extract URL from text
function extractUrl(text: string): string | undefined {
  const match = text.match(/https?:\/\/[^\s)]+/);
  return match ? match[0] : undefined;
}

// Generate stable ID from incident content
function generateId(title: string, linearId?: string): string {
  const key = linearId || title.toLowerCase().replace(/[^a-z0-9]+/g, '-').slice(0, 50);
  return `inc-${key}`;
}

interface ParsedIncident {
  title: string;
  description: string;
  severity: Incident['severity'];
  source: Incident['source'];
  status: Incident['status'];
  linearId?: string;
  url?: string;
  messageId: string;
  timestamp: string;
}

// Try to extract a severity from anywhere in a block of text
function extractSeverity(text: string): Incident['severity'] | null {
  // Match "Severity: MEDIUM-HIGH", "**Severity**: High", "Priority: Urgent", etc.
  const match = text.match(/(?:severity|priority)\s*:?\s*\*{0,2}\s*([A-Za-z][-A-Za-z0-9 ]*)/i);
  if (match) return normalizeSeverity(match[1]);
  return null;
}

// Extract a title from a report — looks for common heading patterns
function extractTitle(text: string): string | null {
  // "Primary Error: ...", "Error: ...", "Issue: ...", "Alert: ..."
  const errorMatch = text.match(/(?:^|\n)\s*[-*]?\s*\*{0,2}(?:Primary Error|Error|Issue|Alert|Problem|Incident)\s*:?\*{0,2}\s*:?\s*(.+)/im);
  if (errorMatch) return errorMatch[1].trim().slice(0, 120);

  // "Service: promotion-service" — use service name as part of title
  const serviceMatch = text.match(/(?:^|\n)\s*[-*]?\s*\*{0,2}Service\*{0,2}\s*:\s*(.+)/im);
  if (serviceMatch) {
    const service = serviceMatch[1].trim();
    // Look for a severity or error description to pair with service
    const severityMatch = text.match(/severity\s*:\s*([A-Za-z][-A-Za-z0-9 ]*)/i);
    const sev = severityMatch ? severityMatch[1].trim() : '';
    return `${service}${sev ? ` — ${sev}` : ''}`.slice(0, 120);
  }

  // "# Title" or "## Title" heading that looks like an incident title
  const headingMatch = text.match(/(?:^|\n)#{1,3}\s+(.+)/m);
  if (headingMatch) {
    const heading = headingMatch[1].trim();
    // Skip generic section headings
    if (!/^(incident summary|code analysis|root cause|timeline|affected|references|recommended|customer impact|investigation|monitoring check|error details|sample error)/i.test(heading)) {
      return heading.slice(0, 120);
    }
  }

  // "1. **Something**: description" — numbered finding
  const findingMatch = text.match(/(?:^|\n)\s*\d+\.\s+\*{2}(.+?)\*{2}/m);
  if (findingMatch) return findingMatch[1].trim().slice(0, 120);

  return null;
}

// Parse a single message for incident patterns
function parseMessageForIncidents(message: Message): ParsedIncident[] {
  const incidents: ParsedIncident[] = [];
  const content = message.content;

  if (!content || content.length < 30) return incidents;

  // Pattern 1: 🎫 [Urgent] Linear ticket EMP-456: Title
  const ticketPattern = /🎫\s*\[(Urgent|High|Medium|Low)\]\s*(?:Linear ticket\s*)?([A-Z]+-\d+):\s*(.+?)(?:\n|$)/gi;
  let match;
  while ((match = ticketPattern.exec(content)) !== null) {
    incidents.push({
      title: match[3].trim(),
      description: content.slice(match.index, match.index + 300),
      severity: normalizeSeverity(match[1]),
      source: 'linear',
      status: parseStatus(content),
      linearId: match[2],
      url: extractUrl(content),
      messageId: message.id,
      timestamp: message.timestamp,
    });
  }

  // Pattern 2: 🚨 NEW High: Title
  const alertPattern = /🚨\s*(?:NEW\s+)?(Urgent|High|Medium|Low):\s*(.+?)(?:\n|$)/gi;
  while ((match = alertPattern.exec(content)) !== null) {
    incidents.push({
      title: match[2].trim(),
      description: content.slice(match.index, match.index + 300),
      severity: normalizeSeverity(match[1]),
      source: parseSource(content),
      status: 'new',
      linearId: extractLinearId(content),
      url: extractUrl(content),
      messageId: message.id,
      timestamp: message.timestamp,
    });
  }

  // Pattern 3: Incident Summary blocks (with any prefix: ###, 🔥, **, or plain text)
  const summaryPattern = /(?:#{1,3}\s*|🔥\s*|\*{2})?\s*Incident Summary\s*\*{0,2}\s*\n([\s\S]*?)(?=\n#{1,3}\s|\n---|\n\*\*\*|$)/gi;
  while ((match = summaryPattern.exec(content)) !== null) {
    const block = match[1];
    const severity = extractSeverity(block);
    const rawSeverity = block.match(/(?:severity|priority)\s*:\s*([A-Za-z][-A-Za-z0-9 ]*)/i);
    // Try multiple title patterns within the summary block
    const titleMatch = block.match(/(?:Primary Error|Title|Issue|Error)\s*:\s*(.+)/i);
    const serviceMatch = block.match(/Service\s*:\s*(.+)/i);
    let title: string | null = null;
    if (titleMatch) {
      title = titleMatch[1].trim();
      // Prepend service name if available
      if (serviceMatch) {
        title = `${serviceMatch[1].trim()}: ${title}`;
      }
    } else if (serviceMatch) {
      const service = serviceMatch[1].trim();
      const sevLabel = rawSeverity ? rawSeverity[1].trim() : '';
      title = sevLabel ? `${service} — ${sevLabel}` : service;
    }
    if (title || severity) {
      incidents.push({
        title: (title || 'Incident detected').slice(0, 120),
        description: block.trim().slice(0, 300),
        severity: severity || 'medium',
        source: parseSource(block),
        status: parseStatus(block),
        linearId: extractLinearId(block),
        url: extractUrl(block),
        messageId: message.id,
        timestamp: message.timestamp,
      });
    }
  }

  // Pattern 4: ✅ Fixed/Resolved status updates
  const resolvedPattern = /✅\s*(?:Fixed|Resolved):\s*(.+?)(?:\n|$)/gi;
  while ((match = resolvedPattern.exec(content)) !== null) {
    incidents.push({
      title: match[1].trim(),
      description: content.slice(match.index, match.index + 300),
      severity: 'medium',
      source: parseSource(content),
      status: 'resolved',
      linearId: extractLinearId(content),
      url: extractUrl(content),
      messageId: message.id,
      timestamp: message.timestamp,
    });
  }

  // Pattern 5: Structured investigation reports with "Root Cause" / "Severity:" / "Customer Impact"
  // This catches the actual firefighter output format
  if (incidents.length === 0) {
    const hasSeverity = extractSeverity(content);
    const hasRootCause = /root cause/i.test(content);
    const hasCodeAnalysis = /code analysis/i.test(content);
    const hasIncidentKeywords = /customer impact|affected systems|deployment issue|investigation|error rate|alert|error count|error details|primary error|incident summary|sample error|service:/i.test(content);

    // If the message looks like an incident report (has severity + at least one structural section)
    if (hasSeverity && (hasRootCause || hasCodeAnalysis || hasIncidentKeywords)) {
      const title = extractTitle(content);
      if (title) {
        incidents.push({
          title,
          description: content.slice(0, 300),
          severity: hasSeverity,
          source: parseSource(content),
          status: parseStatus(content),
          linearId: extractLinearId(content),
          url: extractUrl(content),
          messageId: message.id,
          timestamp: message.timestamp,
        });
      }
    }
  }

  // Pattern 6: Messages with explicit severity markers not caught above
  if (incidents.length === 0) {
    const severity = extractSeverity(content);
    if (severity) {
      const title = extractTitle(content);
      if (title) {
        incidents.push({
          title,
          description: content.slice(0, 300),
          severity,
          source: parseSource(content),
          status: parseStatus(content),
          linearId: extractLinearId(content),
          url: extractUrl(content),
          messageId: message.id,
          timestamp: message.timestamp,
        });
      }
    }
  }

  return incidents;
}

// Merge a parsed incident into the existing incidents map
function mergeIncident(
  existing: Map<string, Incident>,
  parsed: ParsedIncident,
): void {
  const id = generateId(parsed.title, parsed.linearId);
  const prev = existing.get(id);

  if (prev) {
    // Update existing incident
    prev.status = parsed.status;
    prev.lastUpdated = parsed.timestamp;
    if (!prev.messageIds.includes(parsed.messageId)) {
      prev.messageIds.push(parsed.messageId);
    }
    if (parsed.url && !prev.url) prev.url = parsed.url;
    if (parsed.linearId && !prev.linearId) prev.linearId = parsed.linearId;
    // Escalate severity if higher
    const severityOrder: Incident['severity'][] = ['low', 'medium', 'high', 'urgent'];
    if (severityOrder.indexOf(parsed.severity) > severityOrder.indexOf(prev.severity)) {
      prev.severity = parsed.severity;
    }
  } else {
    existing.set(id, {
      id,
      source: parsed.source,
      severity: parsed.severity,
      title: parsed.title,
      description: parsed.description,
      status: parsed.status,
      firstSeen: parsed.timestamp,
      lastUpdated: parsed.timestamp,
      messageIds: [parsed.messageId],
      linearId: parsed.linearId,
      url: parsed.url,
    });
  }
}

export function parseIncidentsFromMessages(messages: Message[]): Incident[] {
  const incidentMap = new Map<string, Incident>();

  for (const message of messages) {
    if (message.role !== 'assistant') continue;

    const parsed = parseMessageForIncidents(message);
    for (const p of parsed) {
      mergeIncident(incidentMap, p);
    }
  }

  // Sort: active first (by severity), then resolved
  const statusOrder: Record<Incident['status'], number> = {
    new: 0, investigating: 1, fixing: 2, testing: 3, resolved: 4, failed: 5,
  };
  const severityOrder: Record<Incident['severity'], number> = {
    urgent: 0, high: 1, medium: 2, low: 3,
  };

  return Array.from(incidentMap.values()).sort((a, b) => {
    const statusDiff = statusOrder[a.status] - statusOrder[b.status];
    if (statusDiff !== 0) return statusDiff;
    return severityOrder[a.severity] - severityOrder[b.severity];
  });
}
