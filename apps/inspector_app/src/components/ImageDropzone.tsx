'use client'

import { useRef, useState, useCallback } from 'react'

interface Props {
  label: string
  file: File | null
  onChange: (file: File | null) => void
  disabled?: boolean
}

export default function ImageDropzone({ label, file, onChange, disabled }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDragging(false)
      if (disabled) return
      const dropped = e.dataTransfer.files[0]
      if (dropped?.type.startsWith('image/')) onChange(dropped)
    },
    [disabled, onChange]
  )

  const previewUrl = file ? URL.createObjectURL(file) : null

  return (
    <div className="flex flex-col gap-1.5">
      <span className="text-xs text-neutral-500">{label}</span>

      <div
        onClick={() => !disabled && inputRef.current?.click()}
        onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
        className={`relative border rounded-lg cursor-pointer overflow-hidden transition-colors
          ${dragging ? 'border-neutral-400 bg-neutral-50' : 'border-neutral-200 bg-neutral-50 hover:border-neutral-300 hover:bg-neutral-100'}
          ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
        `}
      >
        {previewUrl ? (
          <>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={previewUrl}
              alt="preview"
              className="w-full object-cover max-h-64"
              onLoad={() => URL.revokeObjectURL(previewUrl)}
            />
            {!disabled && (
              <div className="absolute inset-0 bg-white/80 opacity-0 hover:opacity-100 transition-opacity flex items-center justify-center">
                <span className="text-xs text-neutral-500">click to replace</span>
              </div>
            )}
          </>
        ) : (
          <div className="flex flex-col items-center justify-center gap-2 py-10 px-4 text-center">
            <svg
              width="16" height="16" viewBox="0 0 24 24" fill="none"
              stroke="currentColor" strokeWidth="1.5" className="text-neutral-300"
            >
              <rect x="3" y="3" width="18" height="18" rx="1" />
              <circle cx="8.5" cy="8.5" r="1.5" />
              <path d="M21 15l-5-5L5 21" />
            </svg>
            <span className="text-xs text-neutral-400">
              {dragging ? 'Drop image' : 'Click or drag to upload'}
            </span>
          </div>
        )}

        <input
          ref={inputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={(e) => onChange(e.target.files?.[0] ?? null)}
          disabled={disabled}
        />
      </div>

      {file && !disabled && (
        <div className="flex items-center justify-between">
          <span className="text-xs text-neutral-400 truncate">{file.name}</span>
          <button
            type="button"
            onClick={() => onChange(null)}
            className="text-xs text-neutral-400 hover:text-neutral-700 transition-colors ml-2 flex-shrink-0"
          >
            remove
          </button>
        </div>
      )}
    </div>
  )
}
