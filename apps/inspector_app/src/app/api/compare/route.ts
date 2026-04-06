import { NextRequest, NextResponse } from 'next/server'

async function getIdToken(audience: string): Promise<string | null> {
  try {
    const res = await fetch(
      `http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=${encodeURIComponent(audience)}`,
      { headers: { 'Metadata-Flavor': 'Google' }, signal: AbortSignal.timeout(2000) }
    )
    if (res.ok) return res.text()
  } catch {
    // Not on GCP -- fine for local dev
  }
  return null
}

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
    })
    const data = await upstream.json()
    return NextResponse.json(data, { status: upstream.status })
  } catch (err) {
    console.error('compare proxy error:', err)
    return NextResponse.json({ error: 'upstream request failed' }, { status: 502 })
  }
}
