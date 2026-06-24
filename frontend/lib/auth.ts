import { apiGet, type User } from "@/lib/api";
import { createClient } from "@/lib/supabase/server";

export async function getCurrentUser(): Promise<User | null> {
  const supabase = await createClient();
  const { data } = await supabase.auth.getSession();
  if (!data.session) return null;

  try {
    const res = await apiGet<User>("/api/v1/users/me", undefined, data.session.access_token);
    return res.data;
  } catch {
    return null;
  }
}
