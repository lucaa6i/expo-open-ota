import { useSearchParams } from 'react-router';
import { useMemo } from 'react';
import { BranchesTable } from '@/pages/Updates/components/BranchesTable';
import { RuntimeVersionsTable } from '@/pages/Updates/components/RuntimeVersionsTable';
import { UpdatesTable } from '@/pages/Updates/components/UpdatesTable';

export const Updates = () => {
  const [searchParams] = useSearchParams();
  const currentBranch = searchParams.get('branch');
  const runtimeVersion = searchParams.get('runtimeVersion');

  const component = useMemo(() => {
    if (!currentBranch) {
      return <BranchesTable />;
    }
    if (!runtimeVersion) {
      return <RuntimeVersionsTable branch={currentBranch} />;
    }
    return <UpdatesTable branch={currentBranch} runtimeVersion={runtimeVersion} />;
  }, [currentBranch, runtimeVersion]);

  return (
    <div className="w-full h-screen flex-1 p-5">
      <h1 className="text-2xl font-medium mb-4">Updates</h1>
      {component}
    </div>
  );
};
