'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Shield, Database, Network, Zap, Globe, Users, Lock, TrendingUp, Settings } from 'lucide-react';

export default function Home() {
  const router = useRouter();
  const [visibleCards, setVisibleCards] = useState<number[]>([]);

  useEffect(() => {
    // Animate cards in sequence
    const timer = setInterval(() => {
      setVisibleCards(prev => {
        if (prev.length < 6) {
          return [...prev, prev.length];
        }
        return prev;
      });
    }, 200);

    return () => clearInterval(timer);
  }, []);

  const features = [
    {
      icon: Shield,
      title: 'Privacy-Preserving',
      description: 'Advanced cryptographic techniques ensure data privacy while enabling accurate matching',
      color: 'text-blue-600',
      bg: 'bg-blue-100'
    },
    {
      icon: Database,
      title: 'Multiple Data Sources',
      description: 'Support for CSV files, PostgreSQL databases, and pre-tokenized datasets',
      color: 'text-green-600',
      bg: 'bg-green-100'
    },
    {
      icon: Network,
      title: 'Secure Networking',
      description: 'TLS encryption, IP whitelisting, and rate limiting for secure data exchange',
      color: 'text-purple-600',
      bg: 'bg-purple-100'
    },
    {
      icon: Zap,
      title: 'High Performance',
      description: 'Optimized algorithms for fast matching across large datasets',
      color: 'text-yellow-600',
      bg: 'bg-yellow-100'
    },
    {
      icon: Globe,
      title: 'Multi-Party Support',
      description: 'Coordinate record linkage across multiple organizations seamlessly',
      color: 'text-indigo-600',
      bg: 'bg-indigo-100'
    },
    {
      icon: Users,
      title: 'Real-World Applications',
      description: 'Healthcare research, fraud detection, and data deduplication use cases',
      color: 'text-red-600',
      bg: 'bg-red-100'
    }
  ];

  const useCases = [
    {
      title: 'Healthcare Research',
      description: 'Link patient records across institutions while maintaining HIPAA compliance',
      icon: '🏥'
    },
    {
      title: 'Fraud Detection',
      description: 'Detect fraudulent activities across financial institutions without sharing sensitive data',
      icon: '🔍'
    },
    {
      title: 'Data Deduplication',
      description: 'Remove duplicate records from merged datasets while preserving privacy',
      icon: '🧹'
    },
    {
      title: 'Population Studies',
      description: 'Conduct large-scale epidemiological research across multiple data sources',
      icon: '📊'
    }
  ];

  const stats = [
    { number: '99.9%', label: 'Privacy Protection', icon: Lock },
    { number: '3+', label: 'Records Processed', icon: Database },
    { number: '99.9%+', label: 'Matching Accuracy', icon: TrendingUp },
    { number: '50ms', label: 'Average Latency', icon: Zap }
  ];

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
      icon: Users,
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
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-blue-50 to-indigo-100">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                <Shield className="h-4 w-4 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-slate-900">CohortBridge</h1>
                <p className="text-xs text-slate-600">By Aluna Labs</p>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <div className="text-sm text-slate-600">
                <span>v1.0.0</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Hero Section */}
        <section className="py-20 text-center">
          <div className="animate-fade-in-up">
            <h1 className="text-5xl font-bold text-slate-900 mb-6">
              Welcome to <span className="text-blue-600">CohortBridge</span>
            </h1>
            <p className="text-xl text-slate-600 mb-8 max-w-3xl mx-auto leading-relaxed">
              The next-generation platform for privacy-preserving record linkage. Connect datasets
              across organizations while maintaining the highest standards of data privacy and security.
            </p>
            <div className="flex justify-center space-x-4">
              <button
                onClick={() => router.push('/get-started')}
                className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition-all transform hover:scale-105 font-medium shadow-lg cursor-pointer"
              >
                Get Started
              </button>
              <button
                onClick={() => window.open('https://github.com/alunalabs/cohort-bridge', '_blank')}
                className="border border-slate-300 text-slate-700 px-8 py-3 rounded-lg hover:bg-slate-50 transition-all transform hover:scale-105 font-medium cursor-pointer"
              >
                View on GitHub
              </button>
            </div>
          </div>
        </section>

        {/* Stats Section */}
        <section className="py-16">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            {stats.map((stat, index) => {
              const IconComponent = stat.icon;
              return (
                <div
                  key={index}
                  className={`text-center p-6 bg-white rounded-xl shadow-lg transform transition-all duration-150 hover:scale-105 ${visibleCards.includes(index)
                    ? 'opacity-100 translate-y-0'
                    : 'opacity-0 translate-y-8'
                    }`}
                  style={{ transitionDelay: `${index * 100}ms` }}
                >
                  <IconComponent className="h-8 w-8 text-blue-600 mx-auto mb-3" />
                  <div className="text-3xl font-bold text-slate-900 mb-2">{stat.number}</div>
                  <div className="text-sm text-slate-600 font-medium">{stat.label}</div>
                </div>
              );
            })}
          </div>
        </section>

        {/* Features Section */}
        <section className="py-20">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-slate-900 mb-4">Why Choose CohortBridge?</h2>
            <p className="text-lg text-slate-600 max-w-2xl mx-auto">
              Built with cutting-edge privacy technologies and designed for real-world applications
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => {
              const IconComponent = feature.icon;
              return (
                <div
                  key={index}
                  className={`bg-white rounded-xl p-8 shadow-lg hover:shadow-xl transition-all duration-150 transform hover:-translate-y-2 ${visibleCards.includes(index)
                    ? 'opacity-100 translate-y-0'
                    : 'opacity-0 translate-y-8'
                    }`}
                  style={{ transitionDelay: `${index * 150}ms` }}
                >
                  <div className={`w-16 h-16 ${feature.bg} rounded-lg flex items-center justify-center mb-6`}>
                    <IconComponent className={`h-8 w-8 ${feature.color}`} />
                  </div>
                  <h3 className="text-xl font-semibold text-slate-900 mb-3">{feature.title}</h3>
                  <p className="text-slate-600 leading-relaxed">{feature.description}</p>
                </div>
              );
            })}
          </div>
        </section>

        {/* How It Works Section */}
        <section className="py-20 bg-white rounded-2xl shadow-xl mb-20">
          <div className="px-8 lg:px-16">
            <div className="text-center mb-16">
              <h2 className="text-4xl font-bold text-slate-900 mb-4">How It Works</h2>
              <p className="text-lg text-slate-600 max-w-2xl mx-auto">
                CohortBridge uses advanced cryptographic techniques to enable secure record linkage
              </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
              <div className="text-center">
                <div className="w-20 h-20 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-6">
                  <span className="text-2xl font-bold text-blue-600">1</span>
                </div>
                <h3 className="text-xl font-semibold text-slate-900 mb-4">Data Tokenization</h3>
                <p className="text-slate-600">
                  Transform sensitive data into privacy-preserving tokens using advanced hashing
                  and encryption techniques while maintaining linkability.
                </p>
              </div>

              <div className="text-center">
                <div className="w-20 h-20 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-6">
                  <span className="text-2xl font-bold text-green-600">2</span>
                </div>
                <h3 className="text-xl font-semibold text-slate-900 mb-4">Secure Matching</h3>
                <p className="text-slate-600">
                  Use probabilistic matching algorithms on tokenized data to identify potential
                  matches without exposing raw personal information.
                </p>
              </div>

              <div className="text-center">
                <div className="w-20 h-20 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-6">
                  <span className="text-2xl font-bold text-purple-600">3</span>
                </div>
                <h3 className="text-xl font-semibold text-slate-900 mb-4">Result Delivery</h3>
                <p className="text-slate-600">
                  Receive match results and confidence scores while maintaining complete privacy
                  of the underlying personal data throughout the process.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Use Cases Section */}
        <section className="py-20">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-slate-900 mb-4">Real-World Applications</h2>
            <p className="text-lg text-slate-600 max-w-2xl mx-auto">
              See how organizations across industries use CohortBridge to unlock insights while protecting privacy
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {useCases.map((useCase, index) => (
              <div
                key={index}
                className="bg-white rounded-xl p-8 shadow-lg hover:shadow-xl transition-all duration-150 transform hover:-translate-y-1"
              >
                <div className="flex items-start space-x-4">
                  <div className="text-4xl">{useCase.icon}</div>
                  <div>
                    <h3 className="text-xl font-semibold text-slate-900 mb-3">{useCase.title}</h3>
                    <p className="text-slate-600 leading-relaxed">{useCase.description}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Configuration Types Section */}
        <section className="py-20">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-slate-900 mb-4">Choose Your Configuration</h2>
            <p className="text-lg text-slate-600 max-w-2xl mx-auto">
              Select the configuration type that best fits your use case and get started in minutes
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {configTypes.map((config) => {
              const IconComponent = config.icon;
              return (
                <div
                  key={config.id}
                  onClick={() => handleConfigSelect(config.path)}
                  className="group relative bg-white rounded-xl shadow-sm border border-slate-200 p-8 hover:shadow-lg hover:border-slate-300 transition-all duration-150 cursor-pointer"
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className={`w-12 h-12 ${config.color} rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-150`}>
                      <IconComponent className="h-6 w-6 text-white" />
                    </div>
                    <div className="opacity-0 group-hover:opacity-100 transition-opacity duration-150">
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
        </section>

        {/* CTA Section */}
        <section className="py-20 text-center">
          <div className="bg-gradient-to-r from-blue-600 to-indigo-600 rounded-2xl p-12 text-white">
            <h2 className="text-3xl font-bold mb-4">Ready to Get Started?</h2>
            <p className="text-xl mb-8 opacity-90">
              Create your first privacy-preserving record linkage configuration in minutes
            </p>
            <button
              onClick={() => router.push('/get-started')}
              className="bg-white text-blue-600 px-8 py-3 rounded-lg hover:bg-gray-100 transition-all transform hover:scale-105 font-medium shadow-lg cursor-pointer"
            >
              Get Started
            </button>
          </div>
        </section>
      </div>

      <style jsx>{`
                @keyframes fade-in-up {
                    from {
                        opacity: 0;
                        transform: translateY(30px);
                    }
                    to {
                        opacity: 1;
                        transform: translateY(0);
                    }
                }
                
                .animate-fade-in-up {
                    animation: fade-in-up 0.8s ease-out;
                }
            `}</style>
    </div>
  );
}
