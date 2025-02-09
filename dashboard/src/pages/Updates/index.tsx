import { api } from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { DataTable } from '@/components/DataTable';

export const Updates = () => {
  const { data, loading, error } = useQuery({
    queryKey: ['data'],
    queryFn: () => api.getReleaseChannels(),
  });

  console.log(data);

  return (
    <div className="w-full h-screen flex-1">
      <div className="w-max min-w-[500px]">
        <DataTable
          columns={[
            {
              id: 'channel',
              header: 'Channel',
              cell: ({ row }) => {
                return row.original;
              },
            },
          ]}
          data={data ?? []}
        />
      </div>
    </div>
  );
};
