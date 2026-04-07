import { NextRequest, NextResponse } from 'next/server'

const TOKEN_COOKIE = 'inspector_token'

export function middleware(request: NextRequest) {
  const appToken = process.env.APP_TOKEN

  // No token configured  open access in dev, blocked in production
  if (!appToken) {
    if (process.env.NODE_ENV !== 'production') return NextResponse.next()
    // Fall through to 401 so misconfigured prod deployments don't silently expose the app
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
  <title>Access Required  Property Inspector</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
      background: #f5f5f4;
      color: #171717;
      min-height: 100vh;
      display: flex;
      align-items: flex-start;
      justify-content: center;
      padding: 3rem 1rem 6rem;
    }
    .wrap { width: 100%; max-width: 36rem; }
    .header { margin-bottom: 1.25rem; padding: 0 0.25rem; }
    .header h1 { font-size: 1rem; font-weight: 600; letter-spacing: -0.01em; }
    .header p { font-size: 0.875rem; color: #a3a3a3; margin-top: 0.25rem; }
    .card {
      background: #fff;
      border: 1px solid #e5e5e5;
      border-radius: 0.75rem;
      box-shadow: 0 1px 3px rgba(0,0,0,0.06);
      padding: 1.25rem;
    }
    .card p { font-size: 0.875rem; color: #525252; line-height: 1.6; }
    .card p + p { margin-top: 0.75rem; }
    code {
      font-family: ui-monospace, 'SF Mono', monospace;
      font-size: 0.8rem;
      background: #f5f5f4;
      border: 1px solid #e5e5e5;
      border-radius: 0.3rem;
      padding: 0.1em 0.4em;
      color: #171717;
    }
    .footer { margin-top: 2rem; padding: 0 0.25rem; font-size: 0.875rem; color: #a3a3a3; }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="header">
      <h1>Property Inspector</h1>
      <p>Access restricted</p>
    </div>
    <div class="card">
      <p>Access to this page is restricted. A valid access token is required.</p>
      <p>Please request a token from the owner and add it to the URL like this:<br>
        <code>?token=YOUR_TOKEN</code>
      </p>
      <p> 
        If you were given a link that already includes a valid token, use that link to access the page.
      </p>
    </div>
    </div>
</body>
</html>`,
    { status: 401, headers: { 'Content-Type': 'text/html' } }
  )
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
}
