/**
 * Fetches a short-lived GCP identity token from the metadata server.
 * Used to authenticate server-side calls to Cloud Run services that require IAM auth.
 * Returns null outside of GCP (local dev) or if the fetch fails for any reason.
 */
export async function getIdToken(audience: string): Promise<string | null> {
  const url =
    `http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity` +
    `?audience=${encodeURIComponent(audience)}&format=full`

  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), 5000)

  try {
    const res = await fetch(url, {
      headers: { 'Metadata-Flavor': 'Google' },
      cache: 'no-store',
      signal: controller.signal,
    })

    if (!res.ok) {
      console.warn(`getIdToken: metadata server returned ${res.status}`)
      return null
    }

    const token = await res.text()
    return token.trim() || null
  } catch (err) {
    // Expected outside of GCP -- log only if we're on GCP (K_SERVICE is set by Cloud Run)
    if (process.env.K_SERVICE) {
      console.error('getIdToken: failed to fetch identity token:', err)
    }
    return null
  } finally {
    clearTimeout(timer)
  }
}
