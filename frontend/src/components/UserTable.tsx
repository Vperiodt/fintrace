import { UserSummary, Pagination } from "../api/types";
import { PaginationControls } from "./PaginationControls";

interface UserTableProps {
  users: UserSummary[];
  loading: boolean;
  error?: string | null;
  pagination?: Pagination;
  onPageChange: (page: number) => void;
  onSelect: (userId: string) => void;
  selectedUserId?: string | null;
}

export function UserTable({
  users,
  loading,
  error,
  pagination,
  onPageChange,
  onSelect,
  selectedUserId
}: UserTableProps) {
  return (
    <div className="table-wrapper">
      {error && <div className="error">{error}</div>}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>User ID</th>
              <th>Name</th>
              <th>Email</th>
              <th>Phone</th>
              <th>KYC Status</th>
              <th>Risk Score</th>
            </tr>
          </thead>
          <tbody>
            {loading && (
              <tr>
                <td colSpan={6} className="placeholder-row">
                  Loading users...
                </td>
              </tr>
            )}
            {!loading && users.length === 0 && (
              <tr>
                <td colSpan={6} className="placeholder-row">
                  No users found.
                </td>
              </tr>
            )}
            {!loading &&
              users.map((user) => (
                <tr
                  key={user.userId}
                  className={user.userId === selectedUserId ? "selected" : ""}
                  onClick={() => onSelect(user.userId)}
                >
                  <td>{user.userId}</td>
                  <td>{user.fullName}</td>
                  <td>{user.email}</td>
                  <td>{user.phone}</td>
                  <td>{user.kycStatus || "-"}</td>
                  <td>{user.riskScore.toFixed(2)}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      {pagination && (
        <PaginationControls
          page={pagination.page}
          totalPages={pagination.totalPages}
          onPageChange={onPageChange}
          disabled={loading}
        />
      )}
    </div>
  );
}
