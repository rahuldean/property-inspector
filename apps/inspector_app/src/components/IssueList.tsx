interface Issue {
  category: string
  severity: string
  description: string
  location: string
  confidence: number
}

const severityStyles: Record<string, {
  row: string
  border: string
  badge: string
  badgeText: string
  label: string
}> = {
  severe: {
    row:       'bg-red-50',
    border:    'border-l-red-400',
    badge:     'bg-red-100',
    badgeText: 'text-red-700',
    label:     'Severe',
  },
  moderate: {
    row:       'bg-amber-50',
    border:    'border-l-amber-400',
    badge:     'bg-amber-100',
    badgeText: 'text-amber-700',
    label:     'Moderate',
  },
  minor: {
    row:       'bg-neutral-50',
    border:    'border-l-neutral-300',
    badge:     'bg-neutral-100',
    badgeText: 'text-neutral-500',
    label:     'Minor',
  },
}

function IssueRow({ issue }: { issue: Issue }) {
  const sev = issue.severity?.toLowerCase() ?? 'minor'
  const s = severityStyles[sev] ?? severityStyles.minor
  const pct = Math.round((issue.confidence ?? 0) * 100)

  return (
    <div className={`rounded-lg border-l-2 px-3.5 py-3 ${s.row} ${s.border}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className={`text-xs font-semibold px-2 py-0.5 rounded-full ${s.badge} ${s.badgeText}`}>
              {s.label}
            </span>
            <span className="text-xs text-neutral-400">{issue.category}</span>
          </div>
          <p className="text-sm text-neutral-800 leading-snug">{issue.description}</p>
          <p className="text-xs text-neutral-400 mt-1">{issue.location}</p>
        </div>
        <span className="text-xs text-neutral-300 font-mono flex-shrink-0 pt-0.5">{pct}%</span>
      </div>
    </div>
  )
}

export default function IssueList({
  issues,
  label,
  emptyMessage = 'None detected.',
}: {
  issues: Issue[]
  label: string
  emptyMessage?: string
}) {
  const severe   = issues.filter(i => i.severity?.toLowerCase() === 'severe')
  const moderate = issues.filter(i => i.severity?.toLowerCase() === 'moderate')
  const minor    = issues.filter(i => i.severity?.toLowerCase() === 'minor')
  const ordered  = [...severe, ...moderate, ...minor]

  return (
    <div>
      <div className="flex items-center gap-3 mb-3">
        <span className="text-xs font-semibold text-neutral-700">{label}</span>
        {severe.length > 0 && (
          <span className="text-xs font-medium text-red-600 bg-red-50 rounded-full px-2 py-0.5">
            {severe.length} severe
          </span>
        )}
        {moderate.length > 0 && (
          <span className="text-xs font-medium text-amber-600 bg-amber-50 rounded-full px-2 py-0.5">
            {moderate.length} moderate
          </span>
        )}
        {minor.length > 0 && (
          <span className="text-xs font-medium text-neutral-500 bg-neutral-100 rounded-full px-2 py-0.5">
            {minor.length} minor
          </span>
        )}
      </div>

      {ordered.length === 0 ? (
        <p className="text-xs text-neutral-400">{emptyMessage}</p>
      ) : (
        <div className="flex flex-col gap-2">
          {ordered.map((issue, i) => (
            <IssueRow key={i} issue={issue} />
          ))}
        </div>
      )}
    </div>
  )
}
