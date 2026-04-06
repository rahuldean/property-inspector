import { NextRequest, NextResponse } from 'next/server'

const TOKEN_COOKIE = 'inspector_token'

export function middleware(request: NextRequest) {
  const appToken = process.env.APP_TOKEN

  // No token configured means dev mode -- open access
  if (!appToken) {
    return NextResponse.next()
  }

  const cookieToken = request.cookies.get(TOKEN_COOKIE)?.value
  if (cookieToken === appToken) {
    return NextResponse.next()
  }

  const queryToken = request.nextUrl.searchParams.get('token')
  if (queryToken === appToken) {
    // Strip the token from the URL, set a session cookie
    const clean = request.nextUrl.clone()
    clean.searchParams.delete('token')
    const response = NextResponse.redirect(clean)
    response.cookies.set(TOKEN_COOKIE, appToken, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 60 * 24 * 7,
      path: '/',
    })
    return response
  }

  return new NextResponse(
    `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Access Required</title>
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: 'Courier New', monospace;
      background: #0f0e0c;
      color: #e8e0d4;
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .container { text-align: center; padding: 2rem; }
    .code { font-size: 5rem; font-weight: 700; color: #f0a030; line-height: 1; }
    .title { font-size: 1.1rem; margin-top: 1rem; color: #8a8070; }
    .hint { font-size: 0.85rem; margin-top: 1.5rem; color: #5a5040; }
    code { color: #f0a030; }
  </style>
</head>
<body>
  <div class="container">
    <div class="code">401</div>
    <div class="title">Access requires a valid token.</div>
    <div class="hint">Append <code>?token=YOUR_TOKEN</code> to the URL.</div>
  </div>
</body>
</html>`,
    { status: 401, headers: { 'Content-Type': 'text/html' } }
  )
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
}
