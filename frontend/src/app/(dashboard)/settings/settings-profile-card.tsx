'use client';

import { Pencil, User as UserIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import type { User } from "@/types";

interface SettingsProfileCardProps {
  user: User | null;
  onEdit: () => void;
}

export function SettingsProfileCard({ user, onEdit }: SettingsProfileCardProps) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-4">
        <div className="flex items-center gap-2">
          <UserIcon className="h-5 w-5 text-primary" />
          <div>
            <CardTitle>My Profile</CardTitle>
            <CardDescription>Manage your personal details</CardDescription>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={onEdit}>
          <Pencil className="mr-2 h-4 w-4" />
          Edit Profile
        </Button>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <Label className="text-xs uppercase text-muted-foreground">Name</Label>
            <p className="font-medium">{user?.name}</p>
          </div>
          <div>
            <Label className="text-xs uppercase text-muted-foreground">Email</Label>
            <p className="font-medium">{user?.email || "Not provided"}</p>
          </div>
          <div>
            <Label className="text-xs uppercase text-muted-foreground">Role</Label>
            <p className="font-medium capitalize">{user?.role_name}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
