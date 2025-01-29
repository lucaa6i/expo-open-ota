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
    title: 'Modular and Flexible',
    description: (
      <>
        Expo Open OTA is built with a modular architecture, allowing you to tailor it to your needs. Host updates on AWS S3 or manage certificates with AWS Secrets Manager, or even host them locally. Every component is designed to be optional, giving you the freedom to adapt the solution to your specific infrastructure.
      </>
    ),
  },
  {
    title: 'Optimized for AWS Integration',
    description: (
      <>
        Expo Open OTA integrates seamlessly with AWS. You can host updates on S3, leverage CloudFront for CDN, and even deploy with Lambda@Edge. While AWS is the primary integration today, the system is designed for extensibility, enabling future support for other platforms.
      </>
    ),
  },
  {
    title: 'Streamlined Adoption with EOAS',
    description: (
      <>
        We developed the EOAS (Expo Open Application Services) NPM package to streamline and automate workflows in CI or local environments. This tool simplifies client-side updates and enables secure publishing through the Expo Open OTA server, reducing friction and accelerating implementation.
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
