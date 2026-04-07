import { NextRequest, NextResponse } from 'next/server'
import { unstable_cache } from 'next/cache'
import { createHash } from 'crypto'
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

  const imageFile = formData.get('image') as File
  const roomName = (formData.get('room_name') as string) ?? 'Unknown Room'
  const floorUnit = (formData.get('floor_unit') as string) ?? ''

  const imageBuffer = Buffer.from(await imageFile.arrayBuffer())
  const imageHash = createHash('sha256').update(imageBuffer).digest('hex')
  const cacheKey = `analyze:${imageHash}:${roomName}:${floorUnit}`

  const headers: Record<string, string> = {}
  const idToken = await getIdToken(apiUrl)
  if (idToken) {
    headers['Authorization'] = `Bearer ${idToken}`
  }

  let cacheMiss = false
  const getCachedResult = unstable_cache(
    async () => {
      cacheMiss = true
      console.log(`[analyze] cache miss — calling upstream (key: ${cacheKey})`)
      const fd = new FormData()
      fd.append('room_name', roomName)
      if (floorUnit) fd.append('floor_unit', floorUnit)
      fd.append('image', new Blob([imageBuffer], { type: imageFile.type || 'image/jpeg' }), imageFile.name)

      const upstream = await fetch(`${apiUrl}/analyze`, {
        method: 'POST',
        headers,
        body: fd,
        cache: 'no-store',
      })
      const data = await upstream.json()
      if (!upstream.ok) throw { status: upstream.status, data }
      console.log(`[analyze] cache filled (key: ${cacheKey})`)
      return data
    },
    [cacheKey],
    { revalidate: 86400 * 7 },
  )

  try {
    const data = await getCachedResult()
    if (!cacheMiss) console.log(`[analyze] cache hit (key: ${cacheKey})`)
    return NextResponse.json(data)
  } catch (err: unknown) {
    const e = err as { status?: number; data?: unknown }
    if (e.status && e.data) {
      return NextResponse.json(e.data, { status: e.status })
    }
    console.error('analyze proxy error:', err)
    return NextResponse.json({ error: 'upstream request failed' }, { status: 502 })
  }
}
