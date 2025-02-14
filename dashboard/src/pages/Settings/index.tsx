import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api.ts';
import { DataTable } from '@/components/DataTable';
import { ApiError } from '@/components/APIError';

export const Settings = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.getSettings(),
  });
  return (
    <div className="w-full h-screen flex-1 p-5">
      <h1 className="text-2xl font-medium mb-4">Settings</h1>
      {!!error && <ApiError error={error} />}
      <DataTable
        columns={[
          {
            header: 'Key',
            accessorKey: 'key',
          },
          {
            header: 'Value',
            accessorKey: 'value',
          },
        ]}
        data={Object.entries(data || {}).map(([key, value]) => ({
          key,
          value,
        }))}
        loading={isLoading}
      />
    </div>
  );
};
