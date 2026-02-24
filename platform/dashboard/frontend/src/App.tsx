import { useState } from 'react';
import { Layout } from './components/Layout';
import { Runs } from './pages/Runs';
import { Costs } from './pages/Costs';
import { Memory } from './pages/Memory';
import { Policies } from './pages/Policies';
import { LiveEvents } from './pages/LiveEvents';

function App() {
  const [page, setPage] = useState('runs');

  const renderPage = () => {
    switch (page) {
      case 'runs': return <Runs />;
      case 'costs': return <Costs />;
      case 'memory': return <Memory />;
      case 'policies': return <Policies />;
      case 'events': return <LiveEvents />;
      default: return <Runs />;
    }
  };

  return (
    <Layout currentPage={page} onNavigate={setPage}>
      {renderPage()}
    </Layout>
  );
}

export default App;
