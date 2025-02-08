import { Layout } from '@/containers/Layout';
import { Route, Routes } from 'react-router';

export const App = () => {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<div />} />
        <Route path="/settings" element={<div>Settings</div>} />
      </Routes>
    </Layout>
  );
};
