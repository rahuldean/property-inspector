const styles: Record<string, string> = {
  excellent: 'text-emerald-700 bg-emerald-50',
  good:      'text-emerald-700 bg-emerald-50',
  fair:      'text-amber-700 bg-amber-50',
  poor:      'text-red-700 bg-red-50',
  unknown:   'text-neutral-500 bg-neutral-100',
}

export default function ConditionBadge({ condition }: { condition: string }) {
  const key = condition?.toLowerCase() ?? 'unknown'
  const style = styles[key] ?? styles.unknown

  return (
    <span className={`inline-flex items-center px-2 py-0.5 text-sm font-medium ${style}`}>
      {condition || 'Unknown'}
    </span>
  )
}
