import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
    {
        title: '⚙️ Ready for production in 10 minutes',
        description: (
            <>
                No database, no complex setup—Expo Open OTA is designed for seamless production use. It fully leverages Expo’s configuration, handling <strong>release channels</strong>, <strong>branches</strong>, and <strong>runtime version fingerprints</strong> out of the box. Just connect your cloud storage (S3) and you're ready to go!
            </>
        ),
    },
    {
        title: '🚀 EOAS: One Command to Publish & Configure',
        description: (
            <>
                Say goodbye to manual setup! Our <code>eoas</code> NPM package automates everything—run <code>npx eoas init</code> to configure your project, and <code>npx eoas publish</code> to push updates effortlessly from your CI/CD pipeline. No extra scripts, no hassle.
            </>
        ),
    },
    {
        title: '⚡ CDN Delivery',
        description: (
            <>
                Your assets, delivered at lightning speed. Expo Open OTA serves static assets through a CDN for maximum performance. Currently supporting AWS CloudFront, with upcoming support for Cloudflare and more—so your users get updates instantly, wherever they are.
            </>
        ),
    },

];


function Feature({title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
