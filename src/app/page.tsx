import { KubeSherlock } from '@/components/kube-sherlock';

export default function Home() {
  return (
    <main className="container mx-auto px-4 py-8 md:py-12">
      <div className="flex flex-col items-center text-center mb-12">
        <div className="mb-4 flex items-center gap-3 text-primary">
          <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" >
            <path d="m9 9-2 2 2 2"/>
            <path d="m13 13 2-2-2-2"/>
            <circle cx="11" cy="11" r="8"/>
            <path d="m21 21-4.3-4.3"/>
          </svg>
          <h1 className="text-4xl md:text-5xl font-bold font-headline tracking-tighter text-foreground">
            Kube Sherlock
          </h1>
        </div>
        <p className="max-w-2xl text-lg text-muted-foreground">
          Your AI-powered assistant for debugging Kubernetes. Just paste an error, and let Sherlock investigate.
        </p>
      </div>
      <KubeSherlock />
    </main>
  );
}
