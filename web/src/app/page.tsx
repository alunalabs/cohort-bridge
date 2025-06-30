'use client';

import { ChevronRight, FileText, ArrowRight, Shield, Database, Activity, Share2, Settings, Zap, Users } from 'lucide-react';
import { useRouter } from 'next/navigation';
import ConfigurationSelector from '../components/ConfigurationSelector';

export default function HomePage() {
  const router = useRouter();

  return (
    <div className="min-h-screen relative">
      {/* Header */}
      <header className="relative z-20 bg-white/95 backdrop-blur-sm border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-6">
          <div className="flex items-center justify-between h-16">
            {/* Logo/Brand */}
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded-lg flex items-center justify-center">
                <Database className="h-5 w-5 text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-gray-900">CohortBridge</h1>
                <p className="text-xs text-gray-500">Configuration Builder</p>
              </div>
            </div>

            {/* Navigation */}
            <nav className="hidden md:flex items-center space-x-8">
              <button
                onClick={() => router.push('/get-started')}
                className="text-gray-600 hover:text-gray-900 transition-colors cursor-pointer"
              >
                Get Started
              </button>
              <button
                onClick={() => router.push('/config/basic')}
                className="text-gray-600 hover:text-gray-900 transition-colors cursor-pointer"
              >
                Configuration Editor
              </button>
              <a
                href="https://github.com/your-org/cohort-bridge"
                target="_blank"
                rel="noopener noreferrer"
                className="text-gray-600 hover:text-gray-900 transition-colors cursor-pointer"
              >
                GitHub
              </a>
            </nav>

            {/* CTA Button */}
            <div className="flex items-center space-x-4">
              <button
                onClick={() => router.push('/config/basic')}
                className="hidden sm:inline-flex items-center px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors font-medium cursor-pointer"
              >
                <Settings className="mr-2 h-4 w-4" />
                Start Building
              </button>

              {/* Mobile menu button */}
              <button className="md:hidden p-2 rounded-lg hover:bg-gray-100 cursor-pointer">
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Fixed animated background - only behind header */}
      <div className="fixed inset-0 pointer-events-none" style={{ zIndex: -1 }}>
        <div className="absolute inset-0 bg-gradient-to-br from-teal-100 via-cyan-50 to-blue-100"></div>

        {/* Glass effect container for the animation with more blur */}
        <div className="absolute inset-0 backdrop-blur-md bg-white/20">
          {/* Record nodes and connections - larger scale and more blurred */}
          <svg className="absolute inset-0 w-full h-full filter blur-sm" viewBox="0 0 1400 800" preserveAspectRatio="xMidYMid slice">
            {/* Connection lines that animate - directional flow */}
            <g className="connections">
              <line x1="100" y1="100" x2="350" y2="120" stroke="url(#gradient1)" strokeWidth="4" opacity="0.8" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="4s" repeatCount="indefinite" />
              </line>
              <line x1="350" y1="120" x2="600" y2="180" stroke="url(#gradient2)" strokeWidth="4" opacity="0.8" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="5s" repeatCount="indefinite" />
              </line>
              <line x1="600" y1="180" x2="850" y2="140" stroke="url(#gradient3)" strokeWidth="4" opacity="0.8" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="4.5s" repeatCount="indefinite" />
              </line>
              <line x1="850" y1="140" x2="1100" y2="110" stroke="url(#gradient1)" strokeWidth="4" opacity="0.8" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="6s" repeatCount="indefinite" />
              </line>
              <line x1="1100" y1="110" x2="1300" y2="90" stroke="url(#gradient2)" strokeWidth="4" opacity="0.8" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="5.5s" repeatCount="indefinite" />
              </line>

              {/* Cross connections - more extensive network */}
              <line x1="150" y1="550" x2="400" y2="480" stroke="url(#gradient2)" strokeWidth="3" opacity="0.6" strokeDasharray="250" strokeDashoffset="250">
                <animate attributeName="stroke-dashoffset" values="250;0;-250" dur="5.5s" repeatCount="indefinite" />
              </line>
              <line x1="400" y1="480" x2="650" y2="600" stroke="url(#gradient3)" strokeWidth="3" opacity="0.6" strokeDasharray="250" strokeDashoffset="250">
                <animate attributeName="stroke-dashoffset" values="250;0;-250" dur="4.8s" repeatCount="indefinite" />
              </line>
              <line x1="650" y1="600" x2="900" y2="520" stroke="url(#gradient1)" strokeWidth="3" opacity="0.6" strokeDasharray="250" strokeDashoffset="250">
                <animate attributeName="stroke-dashoffset" values="250;0;-250" dur="5.2s" repeatCount="indefinite" />
              </line>
              <line x1="900" y1="520" x2="1150" y2="580" stroke="url(#gradient2)" strokeWidth="3" opacity="0.6" strokeDasharray="250" strokeDashoffset="250">
                <animate attributeName="stroke-dashoffset" values="250;0;-250" dur="6.2s" repeatCount="indefinite" />
              </line>

              {/* Vertical connections - spanning full height */}
              <line x1="350" y1="120" x2="400" y2="480" stroke="url(#gradient2)" strokeWidth="2.5" opacity="0.4" strokeDasharray="360" strokeDashoffset="360">
                <animate attributeName="stroke-dashoffset" values="360;0;-360" dur="6.5s" repeatCount="indefinite" />
              </line>
              <line x1="600" y1="180" x2="650" y2="600" stroke="url(#gradient3)" strokeWidth="2.5" opacity="0.4" strokeDasharray="420" strokeDashoffset="420">
                <animate attributeName="stroke-dashoffset" values="420;0;-420" dur="7s" repeatCount="indefinite" />
              </line>
              <line x1="850" y1="140" x2="900" y2="520" stroke="url(#gradient1)" strokeWidth="2.5" opacity="0.4" strokeDasharray="380" strokeDashoffset="380">
                <animate attributeName="stroke-dashoffset" values="380;0;-380" dur="5.8s" repeatCount="indefinite" />
              </line>
              <line x1="1100" y1="110" x2="1150" y2="580" stroke="url(#gradient2)" strokeWidth="2.5" opacity="0.4" strokeDasharray="470" strokeDashoffset="470">
                <animate attributeName="stroke-dashoffset" values="470;0;-470" dur="6.8s" repeatCount="indefinite" />
              </line>

              {/* Diagonal connections for complexity */}
              <line x1="100" y1="100" x2="650" y2="600" stroke="url(#gradient3)" strokeWidth="2" opacity="0.3" strokeDasharray="500" strokeDashoffset="500">
                <animate attributeName="stroke-dashoffset" values="500;0;-500" dur="8s" repeatCount="indefinite" />
              </line>
              <line x1="600" y1="180" x2="1150" y2="580" stroke="url(#gradient1)" strokeWidth="2" opacity="0.3" strokeDasharray="550" strokeDashoffset="550">
                <animate attributeName="stroke-dashoffset" values="550;0;-550" dur="9s" repeatCount="indefinite" />
              </line>

              {/* Additional middle layer connections */}
              <line x1="200" y1="300" x2="500" y2="320" stroke="url(#gradient1)" strokeWidth="3" opacity="0.5" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="5.8s" repeatCount="indefinite" />
              </line>
              <line x1="500" y1="320" x2="800" y2="280" stroke="url(#gradient2)" strokeWidth="3" opacity="0.5" strokeDasharray="300" strokeDashoffset="300">
                <animate attributeName="stroke-dashoffset" values="300;0;-300" dur="6.3s" repeatCount="indefinite" />
              </line>
              <line x1="800" y1="280" x2="1050" y2="350" stroke="url(#gradient3)" strokeWidth="3" opacity="0.5" strokeDasharray="250" strokeDashoffset="250">
                <animate attributeName="stroke-dashoffset" values="250;0;-250" dur="5.1s" repeatCount="indefinite" />
              </line>
            </g>

            {/* Record nodes (data points) - spread vertically */}
            <g className="nodes">
              {/* Top row - Source records */}
              <circle cx="100" cy="100" r="14" fill="#3b82f6" opacity="0.9">
                <animate attributeName="r" values="10;18;10" dur="3s" repeatCount="indefinite" />
              </circle>
              <circle cx="350" cy="120" r="12" fill="#8b5cf6" opacity="0.9">
                <animate attributeName="r" values="8;16;8" dur="3.5s" repeatCount="indefinite" />
              </circle>
              <circle cx="600" cy="180" r="16" fill="#06b6d4" opacity="0.9">
                <animate attributeName="r" values="12;20;12" dur="4s" repeatCount="indefinite" />
              </circle>
              <circle cx="850" cy="140" r="11" fill="#10b981" opacity="0.9">
                <animate attributeName="r" values="7;15;7" dur="2.8s" repeatCount="indefinite" />
              </circle>
              <circle cx="1100" cy="110" r="15" fill="#f59e0b" opacity="0.9">
                <animate attributeName="r" values="11;19;11" dur="3.2s" repeatCount="indefinite" />
              </circle>
              <circle cx="1300" cy="90" r="13" fill="#ef4444" opacity="0.9">
                <animate attributeName="r" values="9;17;9" dur="3.7s" repeatCount="indefinite" />
              </circle>

              {/* Middle layer - Processing records */}
              <circle cx="200" cy="300" r="10" fill="#3b82f6" opacity="0.8">
                <animate attributeName="r" values="6;14;6" dur="4.1s" repeatCount="indefinite" />
              </circle>
              <circle cx="500" cy="320" r="13" fill="#8b5cf6" opacity="0.8">
                <animate attributeName="r" values="9;17;9" dur="3.9s" repeatCount="indefinite" />
              </circle>
              <circle cx="800" cy="280" r="11" fill="#06b6d4" opacity="0.8">
                <animate attributeName="r" values="7;15;7" dur="4.4s" repeatCount="indefinite" />
              </circle>
              <circle cx="1050" cy="350" r="14" fill="#10b981" opacity="0.8">
                <animate attributeName="r" values="10;18;10" dur="3.6s" repeatCount="indefinite" />
              </circle>

              {/* Bottom row - Matched records */}
              <circle cx="150" cy="550" r="12" fill="#3b82f6" opacity="0.8">
                <animate attributeName="r" values="8;16;8" dur="3.8s" repeatCount="indefinite" />
              </circle>
              <circle cx="400" cy="480" r="14" fill="#8b5cf6" opacity="0.8">
                <animate attributeName="r" values="10;18;10" dur="4.2s" repeatCount="indefinite" />
              </circle>
              <circle cx="650" cy="600" r="10" fill="#06b6d4" opacity="0.8">
                <animate attributeName="r" values="6;14;6" dur="3.6s" repeatCount="indefinite" />
              </circle>
              <circle cx="900" cy="520" r="16" fill="#10b981" opacity="0.8">
                <animate attributeName="r" values="12;20;12" dur="4.5s" repeatCount="indefinite" />
              </circle>
              <circle cx="1150" cy="580" r="13" fill="#f59e0b" opacity="0.8">
                <animate attributeName="r" values="9;17;9" dur="4.1s" repeatCount="indefinite" />
              </circle>

              {/* Scattered nodes for full coverage */}
              <circle cx="50" cy="400" r="9" fill="#ec4899" opacity="0.7">
                <animate attributeName="r" values="5;13;5" dur="4.8s" repeatCount="indefinite" />
              </circle>
              <circle cx="1200" cy="420" r="11" fill="#f97316" opacity="0.7">
                <animate attributeName="r" values="7;15;7" dur="3.3s" repeatCount="indefinite" />
              </circle>
              <circle cx="300" cy="200" r="8" fill="#84cc16" opacity="0.6">
                <animate attributeName="r" values="4;12;4" dur="5.2s" repeatCount="indefinite" />
              </circle>
              <circle cx="750" cy="450" r="10" fill="#8b5cf6" opacity="0.6">
                <animate attributeName="r" values="6;14;6" dur="4.6s" repeatCount="indefinite" />
              </circle>
              <circle cx="450" cy="150" r="7" fill="#06b6d4" opacity="0.5">
                <animate attributeName="r" values="3;11;3" dur="5.8s" repeatCount="indefinite" />
              </circle>
              <circle cx="950" cy="380" r="9" fill="#10b981" opacity="0.5">
                <animate attributeName="r" values="5;13;5" dur="4.3s" repeatCount="indefinite" />
              </circle>

              {/* Edge nodes for full coverage */}
              <circle cx="1350" cy="250" r="6" fill="#ec4899" opacity="0.4">
                <animate attributeName="r" values="2;10;2" dur="6.2s" repeatCount="indefinite" />
              </circle>
              <circle cx="25" cy="700" r="8" fill="#f59e0b" opacity="0.4">
                <animate attributeName="r" values="4;12;4" dur="5.5s" repeatCount="indefinite" />
              </circle>
              <circle cx="1250" cy="650" r="7" fill="#3b82f6" opacity="0.4">
                <animate attributeName="r" values="3;11;3" dur="5.9s" repeatCount="indefinite" />
              </circle>
              <circle cx="120" cy="250" r="6" fill="#8b5cf6" opacity="0.4">
                <animate attributeName="r" values="2;10;2" dur="6.5s" repeatCount="indefinite" />
              </circle>
            </g>

            {/* Enhanced gradient definitions for connection lines */}
            <defs>
              <linearGradient id="gradient1" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#3b82f6" stopOpacity="0" />
                <stop offset="50%" stopColor="#3b82f6" stopOpacity="1" />
                <stop offset="100%" stopColor="#3b82f6" stopOpacity="0" />
              </linearGradient>
              <linearGradient id="gradient2" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#8b5cf6" stopOpacity="0" />
                <stop offset="50%" stopColor="#8b5cf6" stopOpacity="1" />
                <stop offset="100%" stopColor="#8b5cf6" stopOpacity="0" />
              </linearGradient>
              <linearGradient id="gradient3" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#06b6d4" stopOpacity="0" />
                <stop offset="50%" stopColor="#06b6d4" stopOpacity="1" />
                <stop offset="100%" stopColor="#06b6d4" stopOpacity="0" />
              </linearGradient>
            </defs>
          </svg>
        </div>
      </div>

      {/* Hero Section - transparent to show animation */}
      <div className="relative overflow-hidden min-h-[600px]">
        <div className="relative max-w-6xl mx-auto px-6 py-16 lg:py-24">
          <div className="text-center max-w-4xl mx-auto">
            {/* Main headline */}
            <h1 className="text-5xl lg:text-6xl font-bold text-gray-900 mb-6 leading-tight">
              Privacy-first patient
              <span className="block text-transparent bg-clip-text bg-gradient-to-r from-blue-500 to-purple-500">
                record linking
              </span>
            </h1>

            {/* Subtitle */}
            <p className="text-xl text-gray-600 mb-8 leading-relaxed max-w-3xl mx-auto">
              Direct peer-to-peer matching between healthcare organizations.
              No middleman, no cloud storage, complete data control.
            </p>

            {/* CTA Buttons */}
            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
              <button
                onClick={() => router.push('/get-started')}
                className="group inline-flex items-center px-6 py-3 bg-blue-500 text-white rounded-xl hover:bg-blue-600 transition-all duration-300 shadow-md hover:shadow-lg font-medium cursor-pointer"
              >
                Get Started
                <ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-1 transition-transform" />
              </button>

              <button
                onClick={() => router.push('/config/basic')}
                className="inline-flex items-center px-6 py-3 text-gray-700 hover:text-gray-900 transition-colors font-medium cursor-pointer"
              >
                Configuration Editor
                <ChevronRight className="ml-1 h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* All remaining sections with white backgrounds */}
      <div className="bg-white relative z-10">

        {/* Features Grid - Clean white boxes */}
        <div className="max-w-6xl mx-auto px-6 py-16">
          <div className="text-center mb-12">
            <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 mb-4">
              Built for healthcare
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Enterprise-grade security meets intuitive design
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {[
              {
                title: "HIPAA Compliant",
                description: "End-to-end encryption with privacy-preserving algorithms",
                icon: Shield,
                color: "bg-blue-500"
              },
              {
                title: "Zero Trust",
                description: "Direct connections with no third-party dependencies",
                icon: Share2,
                color: "bg-green-500"
              },
              {
                title: "Fuzzy Matching",
                description: "Advanced algorithms handle data variations automatically",
                icon: Activity,
                color: "bg-purple-500"
              }
            ].map((feature, index) => (
              <div key={index} className="bg-white rounded-2xl p-8 shadow-sm border border-gray-200 hover:shadow-md transition-shadow group">
                <div className={`inline-flex items-center justify-center w-12 h-12 ${feature.color} rounded-xl mb-6 group-hover:scale-110 transition-transform duration-300`}>
                  <feature.icon className="h-6 w-6 text-white" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">{feature.title}</h3>
                <p className="text-gray-600 leading-relaxed">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>

        {/* Stats Section */}
        <div className="max-w-6xl mx-auto px-6 py-16">
          <div className="bg-white rounded-2xl p-12 shadow-sm border border-gray-200">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
              {[
                { number: "256-bit", label: "AES Encryption", color: "text-blue-500" },
                { number: "99.9%", label: "Match Accuracy", color: "text-green-500" },
                { number: "0", label: "Third Parties", color: "text-purple-500" },
                { number: "100%", label: "Data Control", color: "text-orange-500" }
              ].map((stat, index) => (
                <div key={index} className="group">
                  <div className={`text-3xl lg:text-4xl font-bold mb-2 ${stat.color} group-hover:scale-110 transition-transform duration-300`}>
                    {stat.number}
                  </div>
                  <div className="text-gray-600 font-medium">{stat.label}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* How it works */}
        <div className="max-w-6xl mx-auto px-6 py-16">
          <div className="bg-white rounded-2xl p-12 shadow-sm border border-gray-200">
            <div className="text-center mb-12">
              <h2 className="text-3xl lg:text-4xl font-bold text-gray-900 mb-4">
                How it works
              </h2>
              <p className="text-xl text-gray-600">
                Three simple steps to secure patient matching
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
              {[
                {
                  title: "Configure",
                  description: "Set up your matching parameters using our visual interface",
                  icon: Settings,
                  color: "bg-blue-500"
                },
                {
                  title: "Connect",
                  description: "Establish direct encrypted connections with partner organizations",
                  icon: Zap,
                  color: "bg-green-500"
                },
                {
                  title: "Match",
                  description: "Run privacy-preserving matching while maintaining data control",
                  icon: Users,
                  color: "bg-purple-500"
                }
              ].map((item, index) => (
                <div key={index} className="text-center group">
                  <div className={`inline-flex items-center justify-center w-16 h-16 ${item.color} rounded-2xl mb-6 group-hover:scale-110 transition-transform duration-300`}>
                    <item.icon className="h-8 w-8 text-white" />
                  </div>
                  <h3 className="text-xl font-semibold text-gray-900 mb-4">{item.title}</h3>
                  <p className="text-gray-600 leading-relaxed">{item.description}</p>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Final CTA */}
        <div className="max-w-6xl mx-auto px-6 py-16">
          <div className="bg-gradient-to-r from-blue-500 to-purple-500 rounded-2xl p-12 lg:p-16 text-center text-white">
            <h2 className="text-3xl lg:text-4xl font-bold mb-6">
              Ready to get started?
            </h2>
            <p className="text-xl text-blue-100 mb-8 max-w-2xl mx-auto">
              Create your first configuration and start building secure patient matching infrastructure
            </p>
            <button
              onClick={() => router.push('/get-started')}
              className="inline-flex items-center px-6 py-3 bg-white text-blue-500 rounded-xl hover:bg-gray-100 transition-colors font-semibold shadow-lg cursor-pointer"
            >
              <FileText className="mr-2 h-4 w-4" />
              Start Building
              <ArrowRight className="ml-2 h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Footer */}
        <footer className="border-t border-gray-200 bg-white">
          <div className="max-w-6xl mx-auto px-6 py-8">
            <div className="text-center text-gray-600">
              <p>Built with privacy and security at the core</p>
            </div>
          </div>
        </footer>
      </div>
    </div>
  );
}
