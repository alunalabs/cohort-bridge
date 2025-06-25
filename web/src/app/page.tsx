'use client';

import { useRouter } from 'next/navigation';
import { Settings, Database, Shield, Network, FileText, Download } from 'lucide-react';

export default function Home() {
  const router = useRouter();

  const configTypes = [
    {
      id: 'basic',
      name: 'Basic Configuration',
      description: 'Simple setup for basic record linkage',
      icon: Settings,
      color: 'bg-blue-500',
      path: '/config/basic'
    },
    {
      id: 'postgres',
      name: 'PostgreSQL Configuration',
      description: 'Connect to PostgreSQL databases',
      icon: Database,
      color: 'bg-green-500',
      path: '/config/postgres'
    },
    {
      id: 'secure',
      name: 'Secure Configuration',
      description: 'Enhanced security with timeouts and logging',
      icon: Shield,
      color: 'bg-red-500',
      path: '/config/secure'
    },
    {
      id: 'tokenized',
      name: 'Tokenized Configuration',
      description: 'Work with pre-tokenized data',
      icon: FileText,
      color: 'bg-purple-500',
      path: '/config/tokenized'
    },
    {
      id: 'network',
      name: 'Network Configuration',
      description: 'Multi-party network setup',
      icon: Network,
      color: 'bg-orange-500',
      path: '/config/network'
    }
  ];

  const handleConfigSelect = (path: string) => {
    router.push(path);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-2">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                <Settings className="h-6 w-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-slate-900">CohortBridge</h1>
                <p className="text-sm text-slate-600">By Aluna Labs</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <button
                onClick={() => router.push('/about')}
                className="text-sm text-slate-600 hover:text-slate-900 transition-colors cursor-pointer font-medium"
              >
                About
              </button>
              <div className="text-sm text-slate-600">
                <span>v1.0.0</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="text-center">
          <h2 className="text-4xl font-bold text-slate-900 sm:text-5xl">
            Build Your Configuration
          </h2>
        </div>

        {/* Configuration Types Grid */}
        <div className="mt-16 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {configTypes.map((config) => {
            const IconComponent = config.icon;
            return (
              <div
                key={config.id}
                onClick={() => handleConfigSelect(config.path)}
                className="group relative bg-white rounded-xl shadow-sm border border-slate-200 p-8 hover:shadow-lg hover:border-slate-300 transition-all duration-200 cursor-pointer"
              >
                <div className="flex items-center justify-between mb-4">
                  <div className={`w-12 h-12 ${config.color} rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-200`}>
                    <IconComponent className="h-6 w-6 text-white" />
                  </div>
                  <div className="opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                    <div className="w-6 h-6 text-slate-400">
                      →
                    </div>
                  </div>
                </div>
                <h3 className="text-lg font-semibold text-slate-900 mb-2">
                  {config.name}
                </h3>
                <p className="text-slate-600 text-sm leading-relaxed">
                  {config.description}
                </p>
              </div>
            );
          })}
        </div>

        {/* Features Section */}
        <div className="mt-24 bg-white rounded-2xl shadow-sm border border-slate-200 p-8 md:p-12">
          <div className="text-center mb-12">
            <h3 className="text-2xl font-bold text-slate-900">Why Use Our Builder?</h3>
            <p className="mt-2 text-slate-600">Built for developers, designed for simplicity</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Settings className="h-8 w-8 text-blue-600" />
              </div>
              <h4 className="text-lg font-semibold text-slate-900 mb-2">Visual Interface</h4>
              <p className="text-slate-600 text-sm">No more manual YAML editing. Build configurations with an intuitive form interface.</p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Shield className="h-8 w-8 text-green-600" />
              </div>
              <h4 className="text-lg font-semibold text-slate-900 mb-2">Validation</h4>
              <p className="text-slate-600 text-sm">Built-in validation ensures your configurations are correct before export.</p>
            </div>

            <div className="text-center">
              <div className="w-16 h-16 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <Download className="h-8 w-8 text-purple-600" />
              </div>
              <h4 className="text-lg font-semibold text-slate-900 mb-2">Export Ready</h4>
              <p className="text-slate-600 text-sm">Download production-ready YAML files that work seamlessly with CohortBridge.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
