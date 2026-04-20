export interface Customer {
  id: string;
  name: string;
  email: string;
  created_at: string;
  /** ISO-8601 when soft-deactivated; absent or null when active. */
  deactivated_at?: string | null;
}

export interface CreateCustomerRequest {
  name: string;
  email: string;
}
