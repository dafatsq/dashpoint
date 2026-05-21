export interface AuditLog {
  id: string;
  user_id: string;
  user_name: string;
  action: string;
  entity_type: string;
  entity_id: string;
  description?: string;
  old_values: Record<string, unknown>;
  new_values: Record<string, unknown>;
  ip_address: string;
  user_agent: string;
  created_at: string;
}
