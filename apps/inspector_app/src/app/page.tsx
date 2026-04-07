'use client'

import { useState, useEffect } from 'react'
import ImageDropzone from '@/components/ImageDropzone'
import ConditionBadge from '@/components/ConditionBadge'
import IssueList from '@/components/IssueList'

type Mode = 'analyze' | 'compare'

interface Issue {
  category: string
  severity: string
  description: string
  location: string
  confidence: number
}

interface RoomAnalysis {
  room_meta: { room_name: string; floor_unit: string }
  issues: Issue[]
  summary: string
  overall_condition: string
}

interface ComparisonReport {
  room_meta: { room_name: string; floor_unit: string }
  before_analysis: RoomAnalysis
  after_analysis: RoomAnalysis
  resolved_issues: Issue[]
  new_issues: Issue[]
  unchanged_issues: Issue[]
  summary: string
}

export default function Page() {
  const [mode, setMode] = useState<Mode>('analyze')
  const [roomName, setRoomName] = useState('Living Room')
  const [floorUnit, setFloorUnit] = useState('2A')
  const [image, setImage] = useState<File | null>(null)
  const [analyzeResult, setAnalyzeResult] = useState<RoomAnalysis | null>(null)
  const [before, setBefore] = useState<File | null>(null)
  const [after, setAfter] = useState<File | null>(null)
  const [compareResult, setCompareResult] = useState<ComparisonReport | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function loadTestImages() {
      async function toFile(path: string, name: string) {
        try {
          const res = await fetch(path)
          if (!res.ok) return null
          const blob = await res.blob()
          return new File([blob], name, { type: blob.type || 'image/jpeg' })
        } catch {
          return null
        }
      }

      const [room, beforeImg, afterImg] = await Promise.all([
        toFile('/room.jpg', 'room.jpg'),
        toFile('/before.jpg', 'before.jpg'),
        toFile('/after.jpg', 'after.jpg'),
      ])

      if (room) setImage(room)
      if (beforeImg) setBefore(beforeImg)
      if (afterImg) setAfter(afterImg)
    }

    loadTestImages()
  }, [])

  function switchMode(next: Mode) {
    setMode(next)
    setError(null)
    if (next === 'analyze') {
      setRoomName('Living Room')
      setFloorUnit('2A')
    } else {
      setRoomName('Bed Room')
      setFloorUnit('2A')
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError(null)
    if (mode === 'analyze') {
      setAnalyzeResult(null)
    } else {
      setCompareResult(null)
    }

    const fd = new FormData()
    fd.append('room_name', roomName || 'Unknown Room')
    if (floorUnit) fd.append('floor_unit', floorUnit)

    if (mode === 'analyze') {
      if (!image) return
      fd.append('image', image)
    } else {
      if (!before || !after) return
      fd.append('before', before)
      fd.append('after', after)
    }

    try {
      const endpoint = mode === 'analyze' ? '/api/analyze' : '/api/compare'
      const res = await fetch(endpoint, { method: 'POST', body: fd })
      const data = await res.json()
      if (!res.ok) {
        setError(data.error ?? `Error ${res.status}`)
      } else if (mode === 'analyze') {
        setAnalyzeResult(data)
      } else {
        setCompareResult(data)
      }
    } catch {
      setError('Could not reach the server.')
    } finally {
      setLoading(false)
    }
  }

  const canSubmit = !loading && (mode === 'analyze' ? !!image : !!before && !!after)

  return (
    <div className="min-h-screen bg-neutral-100">
      <div className="max-w-xl mx-auto px-4 pt-12 pb-24">

        {/* Header */}
        <div className="mb-5 px-1">
          <h1 className="text-base font-semibold text-neutral-900 tracking-tight">
            Property Inspector
          </h1>
          <p className="text-sm text-neutral-400 mt-1">
            Sample images are preloaded for illustration - upload your own to inspect a real property and your images are never stored on the server.
          </p>
        </div>

        {/* Form card */}
        <div className="bg-white rounded-xl border border-neutral-200 shadow-sm overflow-hidden">

          {/* Mode tabs */}
          <div className="flex border-b border-neutral-100">
            {(['analyze', 'compare'] as Mode[]).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => switchMode(m)}
                className={`flex-1 py-3 text-sm font-medium transition-colors ${
                  mode === m
                    ? 'text-neutral-900 bg-white'
                    : 'text-neutral-400 bg-neutral-50 hover:text-neutral-600 hover:bg-neutral-50'
                }`}
              >
                {m === 'analyze' ? 'Analyze room' : 'Compare inspections'}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="p-5 flex flex-col gap-4">
            {mode === 'analyze' ? (
              <ImageDropzone label="Photo" file={image} onChange={setImage} disabled={loading} />
            ) : (
              <div className="grid grid-cols-2 gap-3">
                <ImageDropzone label="Before" file={before} onChange={setBefore} disabled={loading} />
                <ImageDropzone label="After"  file={after}  onChange={setAfter}  disabled={loading} />
              </div>
            )}

            <div className="grid grid-cols-2 gap-3">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-neutral-500">Room</label>
                <input
                  type="text"
                  value={roomName}
                  onChange={(e) => setRoomName(e.target.value)}
                  placeholder="Kitchen"
                  disabled={loading}
                  className="bg-neutral-50 border border-neutral-200 rounded-lg px-3 py-2 text-base
                             text-neutral-900 placeholder:text-neutral-300 focus:outline-none
                             focus:border-neutral-400 focus:bg-white transition-colors disabled:opacity-50"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium text-neutral-500">Unit</label>
                <input
                  type="text"
                  value={floorUnit}
                  onChange={(e) => setFloorUnit(e.target.value)}
                  placeholder="2A"
                  disabled={loading}
                  className="bg-neutral-50 border border-neutral-200 rounded-lg px-3 py-2 text-base
                             text-neutral-900 placeholder:text-neutral-300 focus:outline-none
                             focus:border-neutral-400 focus:bg-white transition-colors disabled:opacity-50"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={!canSubmit}
              className="bg-neutral-900 text-white text-base font-medium py-2.5 rounded-lg
                         hover:bg-neutral-700 active:bg-neutral-800 transition-colors
                         disabled:opacity-30 disabled:cursor-not-allowed"
            >
              {loading ? (
                <span className="text-shimmer">
                  {mode === 'analyze' ? 'Analyzing...' : 'Comparing...'}
                </span>
              ) : (
                mode === 'analyze' ? 'Run inspection' : 'Run comparison'
              )}
            </button>

            {error && (
              <p className="text-sm text-red-500 bg-red-50 rounded-lg px-3 py-2">{error}</p>
            )}
          </form>
        </div>

        {/* Analyze results */}
        {mode === 'analyze' && analyzeResult && (
          <div className="mt-4 flex flex-col gap-3">
            {/* Summary card */}
            <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
              <div className="flex items-start justify-between gap-3 mb-3">
                <div>
                  <p className="text-base font-medium text-neutral-900">
                    {analyzeResult.room_meta.room_name}
                    {analyzeResult.room_meta.floor_unit && (
                      <span className="text-neutral-400 font-normal"> · {analyzeResult.room_meta.floor_unit}</span>
                    )}
                  </p>
                  <p className="text-sm text-neutral-400 mt-0.5">
                    {analyzeResult.issues.length} issue{analyzeResult.issues.length !== 1 ? 's' : ''} detected
                  </p>
                </div>
                <ConditionBadge condition={analyzeResult.overall_condition} />
              </div>
              <p className="text-base text-neutral-600 leading-relaxed">{analyzeResult.summary}</p>
            </div>

            {/* Issues card */}
            {analyzeResult.issues.length > 0 && (
              <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
                <IssueList issues={analyzeResult.issues} label="Issues" />
              </div>
            )}
          </div>
        )}

        {/* Compare results */}
        {mode === 'compare' && compareResult && (
          <div className="mt-4 flex flex-col gap-3">
            {/* Summary card */}
            <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
              <div className="flex items-start justify-between gap-3 mb-3">
                <p className="text-base font-medium text-neutral-900">
                  {compareResult.room_meta.room_name}
                  {compareResult.room_meta.floor_unit && (
                    <span className="text-neutral-400 font-normal"> · {compareResult.room_meta.floor_unit}</span>
                  )}
                </p>
                <div className="flex items-center gap-2">
                  <ConditionBadge condition={compareResult.before_analysis.overall_condition} />
                  <span className="text-neutral-300 text-xs">&rarr;</span>
                  <ConditionBadge condition={compareResult.after_analysis.overall_condition} />
                </div>
              </div>
              <p className="text-base text-neutral-600 leading-relaxed">{compareResult.summary}</p>
              <div className="flex gap-3 mt-3 pt-3 border-t border-neutral-100">
                {compareResult.resolved_issues.length > 0 && (
                  <span className="text-sm font-medium text-emerald-600 bg-emerald-50 rounded-full px-2.5 py-0.5">
                    {compareResult.resolved_issues.length} resolved
                  </span>
                )}
                {compareResult.new_issues.length > 0 && (
                  <span className="text-sm font-medium text-red-600 bg-red-50 rounded-full px-2.5 py-0.5">
                    {compareResult.new_issues.length} new
                  </span>
                )}
                {compareResult.unchanged_issues.length > 0 && (
                  <span className="text-sm font-medium text-neutral-500 bg-neutral-100 rounded-full px-2.5 py-0.5">
                    {compareResult.unchanged_issues.length} unchanged
                  </span>
                )}
              </div>
            </div>

            {compareResult.new_issues.length > 0 && (
              <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
                <IssueList issues={compareResult.new_issues} label="New issues" />
              </div>
            )}
            {compareResult.resolved_issues.length > 0 && (
              <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
                <IssueList issues={compareResult.resolved_issues} label="Resolved" emptyMessage="None." />
              </div>
            )}
            {compareResult.unchanged_issues.length > 0 && (
              <div className="bg-white rounded-xl border border-neutral-200 shadow-sm p-5">
                <IssueList issues={compareResult.unchanged_issues} label="Unchanged" emptyMessage="None." />
              </div>
            )}
          </div>
        )}

        <p className="mt-8 px-1 text-sm text-neutral-400">
          Results are AI-generated and should be verified by a qualified inspector.
        </p>
      </div>
    </div>
  )
}
