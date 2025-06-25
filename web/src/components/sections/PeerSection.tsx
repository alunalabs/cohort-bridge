'use client';

import { useFormContext } from 'react-hook-form';
import { Network } from 'lucide-react';

export default function PeerSection() {
    const { register } = useFormContext();

    return (
        <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6">
            <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                    <div className="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
                        <Network className="h-4 w-4 text-green-600" />
                    </div>
                    <h3 className="text-lg font-semibold text-slate-900">Peer Configuration</h3>
                </div>
                <p className="text-sm text-slate-600 ml-11">
                    Set up network connectivity between parties for secure record linkage communication.
                </p>
            </div>

            <div className="space-y-6">
                {/* Peer Host */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Peer Host
                    </label>
                    <input
                        type="text"
                        {...register('peer.host')}
                        placeholder="localhost or 192.168.1.100"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        IP address (e.g., 192.168.1.100) or hostname of the other party in the linkage
                    </p>
                </div>

                {/* Peer Port */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Peer Port
                    </label>
                    <input
                        type="number"
                        {...register('peer.port', { valueAsNumber: true })}
                        placeholder="8081"
                        min="1"
                        max="65535"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Port where the peer system is listening. Coordinate this with the other party.
                    </p>
                </div>

                {/* Listen Port */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Listen Port
                    </label>
                    <input
                        type="number"
                        {...register('listen_port', { valueAsNumber: true })}
                        placeholder="8080"
                        min="1"
                        max="65535"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        Port for incoming connections on this system. Must be available and not blocked by firewall.
                    </p>
                </div>

                {/* Private Key */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">
                        Private Key (Optional)
                    </label>
                    <input
                        type="text"
                        {...register('private_key')}
                        placeholder="Leave empty to auto-generate"
                        className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                    />
                    <p className="mt-1 text-xs text-slate-600">
                        64-character hex key for encryption. Auto-generated if empty. Save this for reuse.
                    </p>
                </div>
            </div>
        </div>
    );
} 