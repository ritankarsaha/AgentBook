import { NextResponse, type NextRequest } from "next/server";
import { createClient } from "@/lib/supabase/server";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function GET(request: NextRequest) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");

  if (!code) {
    return NextResponse.redirect(`${origin}/login?error=missing_code`);
  }

  const supabase = await createClient();
  const { data, error } = await supabase.auth.exchangeCodeForSession(code);

  if (error || !data.session) {
    return NextResponse.redirect(`${origin}/login?error=auth_failed`);
  }

  try {
    await fetch(`${API_URL}/api/v1/users/sync`, {
      method: "POST",
      headers: { Authorization: `Bearer ${data.session.access_token}` },
    });
  } catch {

  }

  return NextResponse.redirect(`${origin}/home`);
}
