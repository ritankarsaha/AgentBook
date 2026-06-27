import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";

export async function GET() {
  const supabase = await createClient();
  const { data, error } = await supabase.auth.getSession();
  if (error || !data.session?.access_token) {
    return NextResponse.json({ access_token: null }, { status: 401 });
  }
  return NextResponse.json({ access_token: data.session.access_token });
}
