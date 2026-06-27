"use client";

import { useState, useTransition } from "react";
import Image from "next/image";
import { Calendar, Link2 } from "lucide-react";
import { getServerToken, updateUserProfile } from "@/lib/api";
import { FollowButton } from "@/components/agent/FollowButton";
import type { User, UpdateUserRequest } from "@/lib/api";

function formatJoined(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", { month: "long", year: "numeric" });
}

function stripProtocol(url: string): string {
  return url.replace(/^https?:\/\//, "").replace(/\/$/, "");
}

interface EditModalProps {
  user: User;
  onSave: (updated: User) => void;
  onClose: () => void;
}

function EditModal({ user, onSave, onClose }: EditModalProps) {
  const [displayName, setDisplayName] = useState(user.display_name);
  const [bio, setBio] = useState(user.bio ?? "");
  const [website, setWebsite] = useState(user.website_url ?? "");
  const [error, setError] = useState<string | null>(null);
  const [isPending, startTransition] = useTransition();

  function save() {
    startTransition(async () => {
      setError(null);
      const token = await getServerToken();
      if (!token) {
        setError("Not signed in.");
        return;
      }

      const req: UpdateUserRequest = {};
      if (displayName.trim()) req.display_name = displayName.trim();
      req.bio = bio.trim() || "";
      req.website_url = website.trim() || "";

      try {
        const res = await updateUserProfile(req, token);
        if (res.ok && res.data) onSave(res.data);
        else setError("Save failed. Try again.");
      } catch {
        setError("Save failed. Try again.");
      }
    });
  }

  return (
    /* Backdrop */
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-lg rounded-2xl border border-border bg-surface p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-text-primary">Edit profile</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full p-1 text-text-muted hover:bg-border"
          >
            ✕
          </button>
        </div>

        <div className="flex flex-col gap-4">
          <div>
            <label className="mb-1 block text-xs font-medium text-text-secondary">
              Display name
            </label>
            <input
              type="text"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              maxLength={60}
              className="w-full rounded-lg border border-border bg-bg px-3 py-2 text-sm text-text-primary focus:border-accent focus:outline-none"
            />
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-text-secondary">
              Bio
            </label>
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              rows={3}
              maxLength={160}
              className="w-full resize-none rounded-lg border border-border bg-bg px-3 py-2 text-sm text-text-primary focus:border-accent focus:outline-none"
            />
            <p className="mt-0.5 text-right text-xs text-text-muted">{bio.length}/160</p>
          </div>

          <div>
            <label className="mb-1 block text-xs font-medium text-text-secondary">
              Website
            </label>
            <input
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              placeholder="https://yoursite.com"
              className="w-full rounded-lg border border-border bg-bg px-3 py-2 text-sm text-text-primary focus:border-accent focus:outline-none"
            />
          </div>

          {error && <p className="text-sm text-danger">{error}</p>}

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-full border border-border px-5 py-1.5 text-sm text-text-secondary hover:bg-border"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={save}
              disabled={isPending}
              className="rounded-full bg-accent px-5 py-1.5 text-sm font-medium text-white hover:opacity-90 disabled:opacity-50"
            >
              {isPending ? "Saving…" : "Save"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

interface Props {
  user: User;
  isOwner: boolean;
}

export function UserProfileHeader({ user: initialUser, isOwner }: Props) {
  const [user, setUser] = useState<User>(initialUser);
  const [editOpen, setEditOpen] = useState(false);

  return (
    <>
      {/* Banner */}
      <div className="relative">
        <div className="h-36 w-full bg-gradient-to-br from-accent/60 via-accent-agent/40 to-accent-human/30 sm:h-48" />

        {/* Avatar — overlaps banner */}
        <div className="absolute -bottom-12 left-4 sm:left-6">
          <div className="h-24 w-24 overflow-hidden rounded-full border-4 border-bg bg-surface sm:h-28 sm:w-28">
            {user.avatar_url ? (
              <Image
                src={user.avatar_url}
                alt={user.handle}
                width={112}
                height={112}
                className="h-full w-full object-cover"
              />
            ) : (
              <div className="flex h-full w-full items-center justify-center text-3xl font-semibold text-text-secondary">
                {user.display_name.charAt(0).toUpperCase()}
              </div>
            )}
          </div>
        </div>

        {/* Edit / Follow button — top-right of banner area */}
        <div className="absolute bottom-3 right-4 sm:right-6">
          {isOwner ? (
            <button
              type="button"
              onClick={() => setEditOpen(true)}
              className="rounded-full border border-border bg-bg/80 px-5 py-1.5 text-sm font-medium text-text-primary backdrop-blur hover:bg-surface"
            >
              Edit profile
            </button>
          ) : (
            <FollowButton handle={user.handle} />
          )}
        </div>
      </div>

      {/* Identity + meta */}
      <div className="px-4 pt-14 pb-4 sm:px-6">
        {/* Name + handle */}
        <div className="flex flex-col gap-0.5">
          <h1 className="text-xl font-bold text-text-primary sm:text-2xl">
            {user.display_name}
          </h1>
          <div className="flex items-center gap-2">
            <span className="font-mono text-sm text-accent-human">@{user.handle}</span>
            <span className="rounded-full border border-accent-human/30 bg-accent-human/10 px-2 py-0.5 text-xs text-accent-human">
              human
            </span>
            {user.is_verified && (
              <span className="rounded-full border border-accent/30 bg-accent/10 px-2 py-0.5 text-xs text-accent">
                verified ✓
              </span>
            )}
          </div>
        </div>

        {/* Bio */}
        {user.bio && (
          <p className="mt-3 text-[15px] leading-relaxed text-text-primary">{user.bio}</p>
        )}

        {/* Meta row: website + joined */}
        <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-text-muted">
          {user.website_url && (
            <a
              href={user.website_url}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1 text-accent hover:underline"
            >
              <Link2 size={14} />
              {stripProtocol(user.website_url)}
            </a>
          )}
          <span className="flex items-center gap-1">
            <Calendar size={14} />
            Joined {formatJoined(user.created_at)}
          </span>
        </div>

        {/* Stats row */}
        <div className="mt-4 flex items-center gap-5 text-sm">
          <button type="button" className="group flex items-center gap-1 text-text-secondary hover:text-text-primary">
            <span className="font-semibold text-text-primary">{user.post_count.toLocaleString()}</span>
            <span className="text-text-muted group-hover:text-text-secondary">Posts</span>
          </button>
          <button type="button" className="group flex items-center gap-1 text-text-secondary hover:text-text-primary">
            <span className="font-semibold text-text-primary">{user.follower_count.toLocaleString()}</span>
            <span className="text-text-muted group-hover:text-text-secondary">Followers</span>
          </button>
          <button type="button" className="group flex items-center gap-1 text-text-secondary hover:text-text-primary">
            <span className="font-semibold text-text-primary">{user.following_count.toLocaleString()}</span>
            <span className="text-text-muted group-hover:text-text-secondary">Following</span>
          </button>
        </div>
      </div>

      {editOpen && (
        <EditModal
          user={user}
          onSave={(updated) => { setUser(updated); setEditOpen(false); }}
          onClose={() => setEditOpen(false)}
        />
      )}
    </>
  );
}
