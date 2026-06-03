export interface Student {
  account: string;
  name: string;
}

export interface ClassInfo {
  class_id: string;
  name: string;
  created_at: string;
}

export interface AttendanceEntry {
  account: string;
  name: string;
  is_present: boolean;
}

export interface ClassDetail {
  class_id: string;
  name: string;
  created_at: string;
  attendance: AttendanceEntry[];
  signin_url: string;
}

export interface LoginResponse {
  token: string;
}

export interface PublicKeyResponse {
  class_id: string;
  public_key: string;
}
