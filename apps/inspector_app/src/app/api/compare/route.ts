import { NextRequest, NextResponse } from 'next/server'
import { getIdToken } from '@/lib/gcp'

export async function POST(request: NextRequest) {
  const apiUrl = process.env.INSPECTOR_API_URL
  if (!apiUrl) {
    return NextResponse.json({ error: 'INSPECTOR_API_URL is not configured' }, { status: 500 })
  }

  let formData: FormData
  try {
    formData = await request.formData()
  } catch {
    return NextResponse.json({ error: 'invalid form data' }, { status: 400 })
  }

  const headers: Record<string, string> = {}
  const idToken = await getIdToken(apiUrl)
  if (idToken) {
    headers['Authorization'] = `Bearer ${idToken}`
  }

  try {
    const upstream = await fetch(`${apiUrl}/compare`, {
      method: 'POST',
      headers,
      body: formData,
      cache: 'no-store',
    })
    const data = await upstream.json()
    return NextResponse.json(data, { status: upstream.status })
  } catch (err) {
    console.error('compare proxy error:', err)
    return NextResponse.json({ error: 'upstream request failed' }, { status: 502 })
  }
}
