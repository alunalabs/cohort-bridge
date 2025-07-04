'use client';

import { Settings, Database, Network, Shield, Clock, FileText, Target, Info, AlertCircle, CheckCircle, Users } from 'lucide-react';
import { useRouter } from 'next/navigation';

const CodeBlock = ({ children, language }: { children: string; language?: string }) => (
    <div className="bg-slate-50 border border-slate-200 rounded p-3 relative overflow-hidden">
        <div className="overflow-x-auto">
            <code className="text-slate-700 text-xs font-mono whitespace-pre block">
                {children}
            </code>
        </div>
    </div>
);

export default function ConfigurationTab() {
    const router = useRouter();

    return (
        <div className="space-y-8">
            <div>
                <h2 className="text-3xl font-bold text-gray-900 mb-4">Understanding CohortBridge Configuration</h2>
                <p className="text-lg text-gray-600 mb-6">
                    CohortBridge uses YAML configuration files to control every aspect of privacy-preserving record linkage.
                    Understanding these settings is crucial for successful deployment and optimal matching performance.
                </p>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                        <Info className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                            <p className="text-sm text-blue-800">
                                <strong>Quick Start:</strong> Use our <a href="/config" className="underline hover:text-blue-900">Configuration Builder</a> to create
                                these files visually, then refer back to this guide to understand what each setting does.
                            </p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="space-y-8">

                {/* Database Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                            <Database className="h-5 w-5 text-blue-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Database Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        The database section defines where your data comes from and how it should be processed for matching.
                        This is the foundation of your entire linkage operation.
                    </p>

                    <div className="space-y-6">
                        <div>
                            <h4 className="font-semibold text-gray-800 mb-3 flex items-center">
                                <CheckCircle className="h-4 w-4 text-green-600 mr-2" />
                                Data Source Type
                            </h4>
                            <div className="bg-gray-50 rounded-lg p-4">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div>
                                        <h5 className="font-medium text-gray-800 mb-2">CSV Files</h5>
                                        <p className="text-sm text-gray-600">
                                            Most common option. Your data is stored in comma-separated value files on disk.
                                            CohortBridge reads these files directly and processes them for matching.
                                        </p>
                                        <ul className="text-xs text-gray-500 mt-2 space-y-1">
                                            <li>• Best for: Small to medium datasets (&lt;1M records)</li>
                                            <li>• Requires: File system access to data files</li>
                                            <li>• Performance: Fast for most use cases</li>
                                        </ul>
                                    </div>
                                    <div>
                                        <h5 className="font-medium text-gray-800 mb-2">PostgreSQL Database</h5>
                                        <p className="text-sm text-gray-600">
                                            Enterprise option. CohortBridge connects directly to your PostgreSQL database
                                            and queries records for processing. Enables real-time linkage workflows.
                                        </p>
                                        <ul className="text-xs text-gray-500 mt-2 space-y-1">
                                            <li>• Best for: Large datasets (&gt;1M records)</li>
                                            <li>• Requires: Database credentials and network access</li>
                                            <li>• Performance: Optimized for large-scale operations</li>
                                        </ul>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-3 flex items-center">
                                <Settings className="h-4 w-4 text-gray-600 mr-2" />
                                Field Mapping & Normalization
                            </h4>
                            <p className="text-sm text-gray-700 mb-3">
                                This is where the magic happens. You tell CohortBridge which columns contain what type of data,
                                and how to standardize that data for optimal matching.
                            </p>
                            <div className="bg-gray-50 rounded-lg p-4">
                                <div className="space-y-4">
                                    <div>
                                        <h6 className="font-medium text-gray-800 mb-2">Why Normalization Matters</h6>
                                        <p className="text-sm text-gray-700 mb-2">
                                            Real-world data is messy. The same person might appear as "Mary Smith" in one database
                                            and "SMITH, MARY" in another. Normalization standardizes these variations so they can match.
                                        </p>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <div className="space-y-3">
                                            <div>
                                                <span className="inline-block bg-blue-100 text-blue-800 text-xs px-2 py-1 rounded font-medium mb-1">name</span>
                                                <p className="text-xs text-gray-600">
                                                    Converts names to lowercase, removes punctuation, normalizes spacing.
                                                    "Mary-Jane O'Connor" becomes "maryjane oconnor"
                                                </p>
                                            </div>
                                            <div>
                                                <span className="inline-block bg-green-100 text-green-800 text-xs px-2 py-1 rounded font-medium mb-1">date</span>
                                                <p className="text-xs text-gray-600">
                                                    Standardizes dates to YYYY-MM-DD format regardless of input format.
                                                    "12/25/1985" becomes "1985-12-25"
                                                </p>
                                            </div>
                                        </div>
                                        <div className="space-y-3">
                                            <div>
                                                <span className="inline-block bg-purple-100 text-purple-800 text-xs px-2 py-1 rounded font-medium mb-1">gender</span>
                                                <p className="text-xs text-gray-600">
                                                    Standardizes gender to single characters.
                                                    "Female" becomes "f", "Male" becomes "m"
                                                </p>
                                            </div>
                                            <div>
                                                <span className="inline-block bg-yellow-100 text-yellow-800 text-xs px-2 py-1 rounded font-medium mb-1">zip</span>
                                                <p className="text-xs text-gray-600">
                                                    Extracts 5-digit ZIP codes from various formats.
                                                    "12345-6789" becomes "12345"
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-3 flex items-center">
                                <Shield className="h-4 w-4 text-red-600 mr-2" />
                                Data Security Options
                            </h4>
                            <div className="space-y-3">
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-2">Tokenized Data</h6>
                                    <p className="text-sm text-gray-700">
                                        CohortBridge will always use tokenized data for matching. Instead of needing to tokenize your data every time, you can use the tokenize command to tokenize your data. If this was done, pass in the tokenized data file.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-2">Encryption</h6>
                                    <p className="text-sm text-gray-700">
                                        If you tokenize your data before, we also offer a way to encrypt the token file. If this was done, simply include the encryption key or a path to the file containing it.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-2">Random Bits (Privacy Enhancement)</h6>
                                    <p className="text-sm text-gray-700">
                                        Adds statistical noise to the matching process. Higher values (0.05-0.1) make it harder
                                        for adversaries to reverse-engineer individual records, but reduce matching accuracy.
                                        Use 0.0 for maximum accuracy in trusted environments.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Network Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
                            <Network className="h-5 w-5 text-green-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Network Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        Network settings control how your CohortBridge instance communicates with other parties.
                        This enables the "peer-to-peer" aspect of the system.
                    </p>

                    <div className="space-y-4">
                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2 flex items-center">
                                <Users className="h-4 w-4 text-blue-600 mr-2" />
                                Peer Connection
                            </h4>
                            <p className="text-sm text-gray-700 mb-3">
                                These settings tell your system where to find the other party you're linking with.
                                Both parties must coordinate these settings for successful connection.
                            </p>
                            <div className="bg-gray-50 rounded-lg p-4">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                    <div>
                                        <h6 className="font-medium text-gray-800 mb-1">Peer Host</h6>
                                        <p className="text-xs text-gray-600">
                                            The IP address or hostname of the other party's system. This could be a local
                                            network address (192.168.1.100) or an internet address (peer.hospital.org).
                                        </p>
                                    </div>
                                    <div>
                                        <h6 className="font-medium text-gray-800 mb-1">Peer Port</h6>
                                        <p className="text-xs text-gray-600">
                                            The specific port number where the other party's CohortBridge is listening.
                                            Common ports are 8080, 8081, or 8443 for secure connections.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Listen Port</h4>
                            <p className="text-sm text-gray-700">
                                The port where your CohortBridge instance will listen for incoming connections from other parties.
                                This must be different from the peer port and must not be blocked by firewalls.
                            </p>
                        </div>

                        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4">
                            <div className="flex items-start space-x-2">
                                <AlertCircle className="h-4 w-4 text-amber-600 mt-0.5" />
                                <div>
                                    <h5 className="font-medium text-amber-900 mb-1">Coordination Required</h5>
                                    <p className="text-sm text-amber-800">
                                        Party A's "peer host:port" must match Party B's "listen port", and vice versa.
                                        If Party A listens on port 8080, Party B must connect to port 8080.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Security Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-red-100 rounded-lg flex items-center justify-center">
                            <Shield className="h-5 w-5 text-red-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Security Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        Security settings protect your system from unauthorized access and abuse. These are critical
                        for production deployments, especially in healthcare and financial environments.
                    </p>

                    <div className="space-y-4">
                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Access Control</h4>
                            <div className="space-y-3">
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">IP Address Restrictions</h6>
                                    <p className="text-sm text-gray-700">
                                        Controls which computers can connect to your CohortBridge instance. You can specify
                                        individual IP addresses (192.168.1.100) or entire subnets (192.168.1.0/24).
                                        This prevents unauthorized parties from attempting connections.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">Rate Limiting</h6>
                                    <p className="text-sm text-gray-700">
                                        Limits how many connection attempts can be made per minute from each IP address.
                                        This prevents denial-of-service attacks and automated scanning tools from
                                        overwhelming your system.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">Connection Limits</h6>
                                    <p className="text-sm text-gray-700">
                                        Controls the maximum number of simultaneous connections. Higher values allow
                                        more concurrent operations but use more system resources. For typical two-party
                                        matching, 3-5 connections is sufficient.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                            <div className="flex items-start space-x-2">
                                <Shield className="h-4 w-4 text-red-600 mt-0.5" />
                                <div>
                                    <h5 className="font-medium text-red-900 mb-1">Production Security</h5>
                                    <p className="text-sm text-red-800">
                                        Always enable IP restrictions and rate limiting for production deployments.
                                        Without these safeguards, your system is vulnerable to unauthorized access and abuse.
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Timeouts Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-yellow-100 rounded-lg flex items-center justify-center">
                            <Clock className="h-5 w-5 text-yellow-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Timeout Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        Timeouts prevent your system from hanging indefinitely when network issues or peer problems occur.
                        Proper timeout configuration ensures reliable operation and faster error detection.
                    </p>

                    <div className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <h4 className="font-semibold text-gray-800 mb-2">Connection Phase</h4>
                                <div className="space-y-2">
                                    <div>
                                        <h6 className="font-medium text-gray-800 text-sm">Connection Timeout</h6>
                                        <p className="text-xs text-gray-600">
                                            How long to wait when establishing the initial TCP connection to a peer.
                                            If this fails, the peer system is likely down or unreachable.
                                        </p>
                                    </div>
                                    <div>
                                        <h6 className="font-medium text-gray-800 text-sm">Handshake Timeout</h6>
                                        <p className="text-xs text-gray-600">
                                            How long to wait for the CohortBridge protocol handshake to complete.
                                            This negotiates encryption and matching parameters between peers.
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <div>
                                <h4 className="font-semibold text-gray-800 mb-2">Data Transfer Phase</h4>
                                <div className="space-y-2">
                                    <div>
                                        <h6 className="font-medium text-gray-800 text-sm">Read Timeout</h6>
                                        <p className="text-xs text-gray-600">
                                            How long to wait when receiving data from a peer. Large datasets require
                                            longer timeouts (120-180 seconds) to prevent premature disconnection.
                                        </p>
                                    </div>
                                    <div>
                                        <h6 className="font-medium text-gray-800 text-sm">Write Timeout</h6>
                                        <p className="text-xs text-gray-600">
                                            How long to wait when sending data to a peer. Should match or exceed
                                            read timeout to account for network asymmetry.
                                        </p>
                                    </div>
                                    <div>
                                        <h6 className="font-medium text-gray-800 text-sm">Idle Timeout</h6>
                                        <p className="text-xs text-gray-600">
                                            How long to keep connections open when no data is being transferred.
                                            Longer timeouts reduce reconnection overhead for multi-stage operations.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                            <h5 className="font-medium text-blue-900 mb-2">Timeout Strategy</h5>
                            <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm text-blue-800">
                                <div>
                                    <h6 className="font-medium">Small Datasets</h6>
                                    <p className="text-xs">Use default values (30-60s) for datasets under 10K records</p>
                                </div>
                                <div>
                                    <h6 className="font-medium">Large Datasets</h6>
                                    <p className="text-xs">Increase read/write to 120-180s for datasets over 100K records</p>
                                </div>
                                <div>
                                    <h6 className="font-medium">Slow Networks</h6>
                                    <p className="text-xs">Double all timeouts when operating over slow or unreliable connections</p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Logging Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
                            <FileText className="h-5 w-5 text-indigo-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Logging Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        Logging captures what happens during matching operations. This is essential for debugging problems,
                        monitoring performance, and meeting compliance requirements in regulated industries.
                    </p>

                    <div className="space-y-4">
                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Log Levels</h4>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                                <div className="text-center p-3 bg-gray-50 rounded">
                                    <div className="font-medium text-gray-800 text-sm">DEBUG</div>
                                    <div className="text-xs text-gray-600 mt-1">Everything that happens</div>
                                    <div className="text-xs text-gray-500 mt-1">Development only</div>
                                </div>
                                <div className="text-center p-3 bg-blue-50 rounded">
                                    <div className="font-medium text-blue-800 text-sm">INFO</div>
                                    <div className="text-xs text-blue-600 mt-1">Important operations</div>
                                    <div className="text-xs text-blue-500 mt-1">Production default</div>
                                </div>
                                <div className="text-center p-3 bg-yellow-50 rounded">
                                    <div className="font-medium text-yellow-800 text-sm">WARN</div>
                                    <div className="text-xs text-yellow-600 mt-1">Potential problems</div>
                                    <div className="text-xs text-yellow-500 mt-1">Issues to monitor</div>
                                </div>
                                <div className="text-center p-3 bg-red-50 rounded">
                                    <div className="font-medium text-red-800 text-sm">ERROR</div>
                                    <div className="text-xs text-red-600 mt-1">Critical failures</div>
                                    <div className="text-xs text-red-500 mt-1">Minimal logging</div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Log Destinations</h4>
                            <div className="space-y-3">
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">File Logging</h6>
                                    <p className="text-sm text-gray-700">
                                        Writes logs to files on disk with automatic rotation when files get too large.
                                        Essential for production systems where you need to review logs later or preserve
                                        them for compliance audits.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">Console Logging</h6>
                                    <p className="text-sm text-gray-700">
                                        Displays logs directly in the terminal where CohortBridge is running.
                                        Useful for development and debugging, but not suitable for production services.
                                    </p>
                                </div>
                                <div>
                                    <h6 className="font-medium text-gray-800 mb-1">System Log (Syslog)</h6>
                                    <p className="text-sm text-gray-700">
                                        Sends logs to the operating system's centralized logging service.
                                        Enables integration with enterprise monitoring tools and log aggregation systems.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Security Audit Logging</h4>
                            <p className="text-sm text-gray-700 mb-2">
                                Creates a separate log file specifically for security events like connection attempts,
                                authentication failures, and access violations. Required for compliance in healthcare
                                (HIPAA) and financial (SOX) environments.
                            </p>
                            <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
                                <div className="flex items-start space-x-2">
                                    <Shield className="h-4 w-4 text-amber-600 mt-0.5" />
                                    <div>
                                        <h6 className="font-medium text-amber-900 text-sm">Compliance Requirement</h6>
                                        <p className="text-xs text-amber-800">
                                            Many regulations require detailed audit trails of who accessed what data and when.
                                            Audit logging provides this automatically.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Matching Configuration Explained */}
                <div className="border border-gray-200 rounded-lg p-6 bg-white">
                    <div className="flex items-center space-x-3 mb-4">
                        <div className="w-8 h-8 bg-emerald-100 rounded-lg flex items-center justify-center">
                            <Target className="h-5 w-5 text-emerald-600" />
                        </div>
                        <h3 className="text-xl font-semibold text-gray-900">Matching Configuration</h3>
                    </div>

                    <p className="text-gray-700 mb-6">
                        Matching thresholds control how similar records need to be before they're considered a match.
                        These settings directly impact the accuracy and completeness of your linkage results.
                    </p>

                    <div className="space-y-4">
                        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
                            <div className="flex items-start space-x-2">
                                <AlertCircle className="h-5 w-5 text-red-600 mt-0.5" />
                                <div>
                                    <h5 className="font-medium text-red-900 mb-1">Critical Requirement</h5>
                                    <p className="text-sm text-red-800">
                                        Both parties must use <strong>identical</strong> threshold values. Different values
                                        will result in zero matches, even for identical records.
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-3">Two-Stage Matching Process</h4>
                            <p className="text-sm text-gray-700 mb-4">
                                CohortBridge uses a two-stage process for maximum efficiency and accuracy. Records must
                                pass both stages to be considered a match.
                            </p>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="border border-blue-200 rounded-lg p-4 bg-blue-50">
                                    <h5 className="font-semibold text-blue-900 mb-2 flex items-center">
                                        <span className="bg-blue-600 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm mr-2">1</span>
                                        Hamming Threshold (Fast Filter)
                                    </h5>
                                    <p className="text-sm text-blue-800 mb-3">
                                        Measures bit-level differences in Bloom filters. This is the first pass that quickly
                                        eliminates obviously non-matching records.
                                    </p>
                                    <div className="space-y-2 text-xs text-blue-700">
                                        <div><strong>Range:</strong> 1-1000 (typically 50-150)</div>
                                        <div><strong>Lower values:</strong> Stricter matching, fewer false positives</div>
                                        <div><strong>Higher values:</strong> More permissive, catches more variations</div>
                                        <div><strong>Default:</strong> 90 (good balance for most use cases)</div>
                                    </div>
                                </div>

                                <div className="border border-green-200 rounded-lg p-4 bg-green-50">
                                    <h5 className="font-semibold text-green-900 mb-2 flex items-center">
                                        <span className="bg-green-600 text-white rounded-full w-6 h-6 flex items-center justify-center text-sm mr-2">2</span>
                                        Jaccard Threshold (Precise Match)
                                    </h5>
                                    <p className="text-sm text-green-800 mb-3">
                                        Measures set similarity using MinHash signatures. This provides the final, precise
                                        similarity score for records that passed the Hamming filter.
                                    </p>
                                    <div className="space-y-2 text-xs text-green-700">
                                        <div><strong>Range:</strong> 0.0-1.0 (0% to 100% similarity)</div>
                                        <div><strong>Higher values:</strong> Stricter matching, fewer false positives</div>
                                        <div><strong>Lower values:</strong> More permissive, catches more variations</div>
                                        <div><strong>Default:</strong> 0.5 (50% similarity required)</div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h4 className="font-semibold text-gray-800 mb-2">Tuning Guidelines</h4>
                            <div className="space-y-3">
                                <div className="bg-gray-50 rounded-lg p-4">
                                    <h6 className="font-medium text-gray-800 mb-2">High Precision (Fewer False Positives)</h6>
                                    <p className="text-sm text-gray-700 mb-2">
                                        Use when false matches are costly and you prefer to miss some true matches rather than
                                        accept incorrect ones.
                                    </p>
                                    <div className="text-xs text-gray-600 font-mono bg-white p-2 rounded">
                                        hamming_threshold: 70, jaccard_threshold: 0.7
                                    </div>
                                </div>

                                <div className="bg-gray-50 rounded-lg p-4">
                                    <h6 className="font-medium text-gray-800 mb-2">High Recall (Fewer False Negatives)</h6>
                                    <p className="text-sm text-gray-700 mb-2">
                                        Use when missing true matches is costly and you're willing to manually review more
                                        potential matches.
                                    </p>
                                    <div className="text-xs text-gray-600 font-mono bg-white p-2 rounded">
                                        hamming_threshold: 120, jaccard_threshold: 0.3
                                    </div>
                                </div>

                                <div className="bg-gray-50 rounded-lg p-4">
                                    <h6 className="font-medium text-gray-800 mb-2">Balanced (Default)</h6>
                                    <p className="text-sm text-gray-700 mb-2">
                                        Good starting point that balances precision and recall for most use cases.
                                    </p>
                                    <div className="text-xs text-gray-600 font-mono bg-white p-2 rounded">
                                        hamming_threshold: 90, jaccard_threshold: 0.5
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            {/* Quick Start Section */}
            <div className="bg-gradient-to-r from-blue-500 to-purple-500 rounded-lg p-6 text-white">
                <h3 className="text-xl font-semibold mb-3">Ready to Configure Your System?</h3>
                <p className="mb-4 opacity-90">
                    Now that you understand what each setting does, use our visual Configuration Builder to create
                    your YAML files quickly and correctly.
                </p>
                <div className="flex flex-col sm:flex-row gap-3">
                    <button
                        onClick={() => router.push('/config')}
                        className="px-6 py-3 bg-white text-blue-600 rounded-lg hover:bg-gray-100 transition-colors font-medium"
                    >
                        Open Configuration Builder
                    </button>
                    <div className="flex items-center space-x-2 text-sm opacity-75">
                        <Info className="h-4 w-4" />
                        <span>Visual interface with validation and examples</span>
                    </div>
                </div>
            </div>
        </div>
    );
} 